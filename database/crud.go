package database

import (
	"context"
	"errors"
	"fmt"
	"log"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func GetDocument(ctx context.Context, dbClient *firestore.Client, docId string, collection string) (*firestore.DocumentSnapshot, error) {
	dsnap, err := dbClient.Collection(collection).Doc(docId).Get(ctx)
	if !dsnap.Exists() {
		log.Printf("Document Doesn't Exists: %v", docId)
		return dsnap, errors.New("Document Doesn't Exists")
	}
	if err != nil {
		log.Printf("Failed Getting document: %v", err)
		return dsnap, err
	}
	return dsnap, nil
}

// CheckUser Function checks if user exists or not, if doesn't then returns nil otherwise returns error
func CheckUser(ctx context.Context, dbClient *firestore.Client, id string) error {
	if id == "" {
		return errors.New("Invalid user id")
	}
	_, err := GetDocument(ctx, dbClient, id, UserCollectionName)
	if err.Error() == "Document Doesn't Exists" {
		return nil
	}
	return err
}

// Get User, retrieves the user information from the DB
func GetUser(ctx context.Context, dbClient *firestore.Client, id string) (User, error) {
	var u User
	if id == "" {
		return u, errors.New("Invalid user id")
	}
	dsnap, err := GetDocument(ctx, dbClient, id, UserCollectionName)
	if err != nil {
		return u, err
	}
	dsnap.DataTo(&u)
	return u, nil
}

func UpdateDocument(ctx context.Context, dbClient *firestore.Client, docId string, field string, value interface{}, collection string) error {
	if docId == "" || field == "" || collection == "" {
		return errors.New("Invalid user id or field value or collection name")
	}
	_, err := dbClient.Collection(collection).Doc(docId).Update(ctx, []firestore.Update{
		{
			Path:  field,
			Value: value,
		},
	})
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred in UpdateDocument: %s", err)
	}
	return err
}

func UpdateUser(ctx context.Context, dbClient *firestore.Client, id string, field string, value interface{}) error {
	if id == "" || field == "" {
		return errors.New("Invalid user id or field value")
	}
	err := UpdateDocument(ctx, dbClient, id, field, value, UserCollectionName)
	return err
}

func SaveDocument(ctx context.Context, dbClient *firestore.Client, docId string, data interface{}, collection string) error {
	if docId == "" {
		return errors.New("Invalid document id")
	}
	_, err := dbClient.Collection(collection).Doc(docId).Set(ctx, data)
	if err != nil {
		log.Printf("Failed Saving data: %v", err)
	}
	log.Printf("Saved data successfully: %v", err)
	return err
}

func SaveUser(ctx context.Context, dbClient *firestore.Client, id string, user User) error {
	if id == "" || id != user.Id {
		return errors.New("Invalid user id")
	}
	err := SaveDocument(ctx, dbClient, id, user, UserCollectionName)
	return err
}

func AddRoom(ctx context.Context, dbClient *firestore.Client, id string, room Room) error {
	if id == "" || room.RoomId == "" {
		return errors.New("Invalid user id or room id")
	}
	var user User
	// get the user document
	user, err := GetUser(ctx, dbClient, id)
	if err != nil {
		// log.Printf("User Doesn't Exists: %v", id)
		return err
	}
	// append the new room
	user.Rooms = append(user.Rooms, room)
	// update the user
	err = UpdateUser(ctx, dbClient, id, "Rooms", user.Rooms)
	if err != nil {
		return err
	}
	log.Printf("Room saved successfully for user: %v", id)
	return nil
}

// Delete Room
func DeleteRoom(ctx context.Context, dbClient *firestore.Client, id string, roomId string) error {
	if id == "" || roomId == "" {
		return errors.New("Invalid user id or room id")
	}
	var user User
	// get the user document
	user, err := GetUser(ctx, dbClient, id)
	if err != nil {
		// log.Printf("User Doesn't Exists: %v", id)
		return err
	}
	// iterate through the array of rooms and find that roomId and delete it
	var roomIndx int = -1
	for i, room := range user.Rooms {
		if room.RoomId == roomId {
			roomIndx = i
			break
		}
	}
	if roomIndx == -1 {
		log.Printf("Room %v not present for user: %v", roomId, id)
		return fmt.Errorf("Room %v not present for user: %v", roomId, id)
	}
	user.Rooms[roomIndx] = user.Rooms[len(user.Rooms)-1]
	user.Rooms = user.Rooms[:len(user.Rooms)-1]
	// save the user
	err = UpdateUser(ctx, dbClient, id, "Rooms", user.Rooms)
	if err != nil {
		return err
	}
	fmt.Printf("Room %s deleted succefully for user: %s", roomId, id)
	return err
}

func DeleteDocument(ctx context.Context, dbClient *firestore.Client, docId string, collection string) error {
	if docId == "" || collection == "" {
		return errors.New("Invalid document id or collection name")
	}
	_, err := dbClient.Collection(collection).Doc(docId).Delete(ctx)
	if err != nil {
		// Handle any errors in an appropriate way, such as returning them.
		log.Printf("An error has occurred in DeleteUser: %s", err)
	}
	return err
}

// Delete User
func DeleteUser(ctx context.Context, dbClient *firestore.Client, id string) error {
	if id == "" {
		return errors.New("Invalid user id")
	}
	err := DeleteDocument(ctx, dbClient, id, UserCollectionName)
	if err != nil {
		return err
	}
	log.Printf("User %s Deleted Successfully", id)
	return err
}

func DeleteField(ctx context.Context, dbClient *firestore.Client, id string, field string) error {
	if id == "" || field == "" {
		return errors.New("Invalid user id or field value")
	}
	err := UpdateUser(ctx, dbClient, id, field, firestore.Delete)
	if err != nil {
		return err
	}
	log.Printf("Field %s Deleted Successfully for User: %s", field, id)
	return err
}

func RunWhereQuery(ctx context.Context, dbClient *firestore.Client, field string, collection string, query string, compareData interface{}) *firestore.DocumentIterator {
	iter := dbClient.Collection(collection).Where(field, query, compareData).Documents(ctx)
	return iter
}

func CheckRoom(ctx context.Context, dbClient *firestore.Client, roomId string) (string, error) {
	if roomId == "" {
		return "", errors.New("Invalid room id")
	}
	compareData := []Room{
		{
			RoomId:   roomId,
			IsLocked: true,
		},
		{
			RoomId:   roomId,
			IsLocked: false,
		},
	}
	iter := RunWhereQuery(ctx, dbClient, "Rooms", UserCollectionName, "array-contains-any", compareData)
	doc, err := iter.Next()
	if err == iterator.Done {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("Error In CheckRoom : %v", err)

	}
	
	return doc.Ref.ID, fmt.Errorf("Room already present: %s", roomId)
}

// New Room
func NewRoom(ctx context.Context, dbClient *firestore.Client, id string, room Room) error {
	if id == "" || room.RoomId == "" {
		return errors.New("Invalid user or room id")
	}
	_, err := CheckRoom(ctx, dbClient, room.RoomId)
	if err != nil {
		return err
	}
	err = AddRoom(ctx, dbClient, id, room)
	if err != nil {
		return err
	}
	fmt.Printf("Room added successfully to user: %s", id)
	return nil
}

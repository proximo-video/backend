package signaling

// TYPES

// offer type
const OFFER string = "OFFER"

// answer type
const ANSWER string = "ANSWER"

// candidate type
const CANDIDATE string = "CANDIDATE"

/*----------------------------------------------------------*/

// REQUEST ACTIONS

// to start a new meeting room
const START string = "START"

// to join room
const JOIN string = "JOIN"

// to end room / meeting - only by "owner"
const END string = "END" // *also a reply message

// to leave the room / meeting
const LEAVE string = "LEAVE"

// to send any message to room memebers
const MESSAGE string = "MESSAGE"

// to approve enterance of a user in room
const APPROVE string = "APPROVE"

// to reject enterance of a user in room
const REJECT string = "REJECT"

/*----------------------------------------------------------*/

// REPLY ACTIONS

// action to indicate send "offer"
const READY string = "READY"

// action to indicate error
const ERROR string = "ERROR"

// action for approval request for entrance into room
const PERMIT string = "PERMIT"

// action to indicate wait
const WAIT string = "WAIT"

/*----------------------------------------------------------*/

// UNREGISTER ACTIONS

// action to ungister all users i.e. end meeting
const ALL string = "ALL"

// action to unregister just one user
const SELF string = "SELF"

package communication

// CommunicationOperation defines strongly-typed constants for permitted operations.
type CommunicationOperation string

const (
	OperationSendMessage   CommunicationOperation = "send_message"
	OperationReceiveMessage CommunicationOperation = "receive_message"
	OperationGetHistory    CommunicationOperation = "get_history"
	OperationDeleteMessage CommunicationOperation = "delete_message"
	OperationMarkRead      CommunicationOperation = "mark_read"
	OperationMarkUnread    CommunicationOperation = "mark_unread"
	OperationGetStatus     CommunicationOperation = "get_status"
)

// IsValid validates if a string matches a known CommunicationOperation.
func (o CommunicationOperation) IsValid() bool {
	switch o {
	case OperationSendMessage, OperationReceiveMessage, OperationGetHistory,
		OperationDeleteMessage, OperationMarkRead, OperationMarkUnread,
		OperationGetStatus:
		return true
	}
	return false
}

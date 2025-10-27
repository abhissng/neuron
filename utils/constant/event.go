package constant

// constants for common messages or events
const (
	// General messages
	SuccessMessage = "Success"
	ErrorMessage   = "Error"
	WarningMessage = "Warning"
	InfoMessage    = "Info"
	DebugMessage   = "Debug"

	// Adapter related messages
	AdapterLoaded     = "AdapterLoaded"
	AdapterUnloaded   = "AdapterUnloaded"
	AdapterInitialize = "AdapterInitialize"
	AdapterStart      = "AdapterStart"
	AdapterStop       = "AdapterStop"
	AdapterError      = "AdapterError"

	// Transaction related messages
	TransactionStarted = "TransactionStarted"
	TransactionSuccess = "TransactionSuccessful"
	TransactionFailed  = "TransactionFailed"
	TransactionPending = "TransactionPending"

	// Data related messages
	DataReceived  = "DataReceived"
	DataProcessed = "DataProcessed"
	DataStored    = "DataStored"
	DataRetrieved = "DataRetrieved"
	DataNotFound  = "DataNotFound"
	DataInvalid   = "DataInvalid"

	// System related messages
	SystemStarted = "SystemStarted"
	SystemStopped = "SystemStopped"
	SystemReady   = "SystemReady"
	SystemError   = "System Error"
	SystemWarning = "System Warning"
	LibraryError  = "Library Error"

	// Event related messages
	EventPublished                   = "EventPublished"
	EventPublishedFailed             = "EventPublishedFailed"
	EventReceived                    = "EventReceived"
	EventProcessed                   = "EventProcessed"
	EventFailed                      = "EventFailed"
	EventHandled                     = "EventHandled"
	SubjectSubscribed                = "SubjectSubscribed"
	SubjectUnsubscribed              = "SubjectUnsubscribed"
	SubjectWithQueueSubscribed       = "SubjectWithQueueSubscribed"
	SubjectWithQueueUnsubscribed     = "SubjectWithQueueUnsubscribed"
	SubjectWithQueueSubscribedFailed = "SubjectWithQueueSubscribedFailed"
	SubjectSubscribeFailed           = "SubjectSubscribeFailed"
	SubscribeSyncFailed              = "SubscribeSyncFailed"
	QueueSubscribeSyncFailed         = "QueueSubscribeSyncFailed"
	MessageProcessed                 = "MessageProcessed"
	ProcessingFailed                 = "ProcessingFailed"
	ConnectionClosed                 = "ConnectionClosed"
	ConnectionClosing                = "ConnectionClosing"
	Publish                          = "Publish"
	Subscribe                        = "Subscribe"
	StartServiceSuccessful           = "StartServiceSuccessful"
	StartServiceFailed               = "StartServiceFailed"

	// Handler related messages
	HandlerStarted    = "HandlerStarted"
	HandlerSuccess    = "HandlerSuccessful"
	HandlerFailed     = "HandlerFailed"
	HandlerPending    = "HandlerPending"
	HandlerRedirect   = "HandlerRedirect"
	MiddlewareSuccess = "MiddlewareSuccessful"
	MiddlewareFailed  = "MiddlewareFailed"

	TransactionMessage    = "Transaction Message"
	AdaptersMessage       = "Adapter Message"
	APICallMessage        = "API Call Message"
	CommunicatorMessage   = "Communicator Message"
	CommandMessage        = "Command Message"
	EventMessage          = "Event Message"
	QueryMessage          = "Query Message"
	NotificationMessage   = "Notification Message"
	HeartbeatMessage      = "Heartbeat Message"
	ControllerMessage     = "Controller Message"
	ServiceHandlerMessage = "Service Handler Message"
	EngineHandlerMessage  = "Engine Handler Message"
	// StartServiceMessage   = "Start Service Message"
)

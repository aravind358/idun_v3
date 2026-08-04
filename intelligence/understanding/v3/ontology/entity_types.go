package ontology

// EntityType defines the canonical semantic type of an extracted entity.
type EntityType string

const (
	// People
	EntityPerson       EntityType = "PERSON"
	EntityContact      EntityType = "CONTACT"
	EntityOrganization EntityType = "ORGANIZATION"
	EntityRole         EntityType = "ROLE"
	EntityProfession   EntityType = "PROFESSION"

	// Geography
	EntityLocation EntityType = "LOCATION"
	EntityCountry  EntityType = "COUNTRY"
	EntityState    EntityType = "STATE"
	EntityCity     EntityType = "CITY"
	EntityAddress  EntityType = "ADDRESS"
	EntityBuilding EntityType = "BUILDING"

	// Files
	EntityFile       EntityType = "FILE"
	EntityDirectory  EntityType = "DIRECTORY"
	EntityDocument   EntityType = "DOCUMENT"
	EntityArchive    EntityType = "ARCHIVE"
	EntityExecutable EntityType = "EXECUTABLE"

	// Computer / Systems
	EntityApplication    EntityType = "APPLICATION"
	EntityProcess        EntityType = "PROCESS"
	EntityService        EntityType = "SERVICE"
	EntityCommand        EntityType = "COMMAND"
	EntityDevice         EntityType = "DEVICE"
	EntitySystemResource EntityType = "SYSTEM_RESOURCE"

	// Numbers
	EntityNumber     EntityType = "NUMBER"
	EntityQuantity   EntityType = "QUANTITY"
	EntityUnit       EntityType = "UNIT"
	EntityPercentage EntityType = "PERCENTAGE"
	EntityCurrency   EntityType = "CURRENCY"

	// Physical
	EntityProduct  EntityType = "PRODUCT"
	EntityFood     EntityType = "FOOD"
	EntityTool     EntityType = "TOOL"
	EntityVehicle  EntityType = "VEHICLE"
	EntityMedicine EntityType = "MEDICINE"

	// Communication
	EntityMessage      EntityType = "MESSAGE"
	EntityConversation EntityType = "CONVERSATION"
	EntityPhoneNumber  EntityType = "PHONE_NUMBER"
	EntityEmail        EntityType = "EMAIL"

	// Internet
	EntityURL       EntityType = "URL"
	EntityWebsite   EntityType = "WEBSITE"
	EntityDomain    EntityType = "DOMAIN"
	EntityIPAddress EntityType = "IP_ADDRESS"

	// AI
	EntityPrompt     EntityType = "PROMPT"
	EntityModel      EntityType = "MODEL"
	EntityToolCall   EntityType = "TOOL_CALL"
	EntityCapability EntityType = "CAPABILITY"
	EntityWorkflow   EntityType = "WORKFLOW"
	EntityAgent      EntityType = "AGENT"

	// Generic
	EntityUnknown    EntityType = "UNKNOWN"
	EntityIdentifier EntityType = "IDENTIFIER"
	EntityLabel      EntityType = "LABEL"
	EntityTag        EntityType = "TAG"
	EntityMetadata   EntityType = "METADATA"
)

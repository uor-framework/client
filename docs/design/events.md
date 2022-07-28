# Events

An Event is the entire interaction of the UOR Client with a referenced Collection. Since a Collection can have many nested Collections, a single event is the complete interaction with all Collections referenced by the UOR Client.

Collections have embedded Event Engines. An Event Engine contains the application logic of a Collection. 

## Event Engine Lookup

1. UOR Client retrieves Collection root node
2. UOR Client looks up Schema in Collection root node
3. UOR Client retrieves the Schema Collection root node
4. UOR Client looks up application logic in the Schema Collection root node
5. UOR Client retrieves the application logic collection root node
6. UOR Client looks up application logic by platform/arch
7. UOR Client retrieves the application logic by platform/arch


## Router

When the UOR Client pushes or pulls (an object), it performs a routing action. The UOR Client can perform 2 routing actions:

1. Route object to Collection
2. Route object to Event Engine


### Router API
| Control Message   |
|-------------------|
| attributes (map)  | The attributes of the object
| digest            | The digest of the object
| source            | The source address of the object
| destination       | The destination address of the object
| payload (bool)    | If no payload, the object is only attributes

# Event Engine Design

Event Engines are designed as follows:
1. Event Engines can be UOR native applications or legacy applications ported to UOR.
2. An application becomes an Event Engine when it implements the UOR Router API and performs tasks pertinent to a UOR Collection. 
3. Event Engines receive application control signalling encoded in attributes. 
4. UOR Native Applications implement the UOR local cache structure, while UOR ported applications rely on URL encoded attribute maps to reference resources.

Event Engine's follow the following workflow:
1. Event Engine process spawned by the UOR Client
2. UOR Client sends Schema default object and the control message to the Event Engine
3. Prepends the source to all attributes in the Control Message
4. Incoming objects from the router are streamed into the needed format
5. Event engine application execution parameters are converted from attributes
6. Event Engine application is executed with parameters

Additionally; 
- Event Engines may be called without passing an object payload to an Event Engine.
- The user context is considered to be a Collection with an Event Engine. Objects are output to the user's context by the UOR Router when signaled. 








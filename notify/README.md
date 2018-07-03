# Notify Protocol

The Notify protocol is used to distribute and sync information between the Edge-Cloud services of the Controller, CRM-Front/Back, and DME.

# Connectivity

Connectivity between services is organized as a tree, with Controllers at the top, CRM-Fronts below them, and CRM-Back and DMEs below that. Connectivity is established bottom-up, so CRM-Back/DMEs have a list (or subset list) of CRM-Fronts above them, and connect to one of them. CRM-Fronts will have a list (or subset list) of Controllers above them, and will connect to one of them. A node may have a single node it connects to above it, and/or may have multiple nodes connecting it from below. Nodes never connect horizontally, i.e. CRM-Front will never connect to another CRM-Front.

This connectivity is established by the notify protocol. Once connected and an initial negotation is done, the connection is fully bidirectional and both directions are independent.

The notify protocol itself does not care about the number of levels in the tree, nor the radix. It does have some optimizations of what type of data to push based on node type, but other than that it doesn't care about the node type either.

# Data Mirroring

The server side (upstream) pushes objects from the persistent (etcd) database down to clients. The client pushes objects containing dynamic state (cloudlet resources, etc) up to upstream server nodes.

In effect, both directions are mirroring their local state onto the remote node. The goal is to keep the remote state in sync with the local state. Pushes are triggered by node-specific code calling into the notify code to tell it about the key of an object that has been changed/deleted. The notify code will then look up a copy of that data and then push it to the remote node.

There is a difference in behavior between upstream/downstream when handling disconnects. When a client node loses a server, it tries to reconnect to a new server. In the meantime, it does not change its copy of the upstream data. When it manages to reconnect, the server will resend it the full data. Once the initial send is done, the client removes any data that was not received (because that data had been removed while it was disconnected). This minimizes changes to the local mirrored data in the face of server disconnects. However, when a server loses its client, it flushes all data related to that client. This is because the client may reconnect to a different server, or may just be gone forever. So the notify code tracks downstream data on a per-client basis using a NotifyId.

# Bandwidth efficient object recognition for drone swarms

### DESCRIPTION:
The aim of this project is to improve object detection accuracy by taking advantage of multiple drones' viewpoints.
An host must be setup in order to receive predictions from the nodes of the swarm (baseStation folder)
The hosts of the swarm need to run the IPCHandler code

### RESULTS:
Our system is able to improve detection accuracy by eliminating some false positives and false negatives. Detailed comparisons, in terms of precision and recall, between our multi-host system and a single-host system are shown in the report.

### STRUCTURE:
    - IPCHandler/ : folder containing the inter process communication handler (for getting data from the python process through UNIX socket). It must be executed on the nodes of the network.
    - structures/ : folder containing structures and functions that handle the communication with other hosts
    - baseStation/ : folder containing the code for executing the base station

### TO RUN THE CODE:
    1. Install Go: https://golang.org/dl/
    2. Install dependencies: go get github.com/dedis/protobuf
    3. Run base station to receive rounds' predictions: baseStation/baseStation --address=<IP:PORT>, IP:PORT indicate the address for reaching the base station
    4. Run the hosts in the swarm: IPChandler/IPChandler
    
### Execution sample (3 nodes, local network and local base station):
    - baseStation/baseStation -address=127.0.0.1:3000
    - IPChandler/IPChandler -nodeID=1 -address=127.0.0.1:5001 -peers=127.0.0.1:5002,127.0.0.1:5003 -class=person -BS=127.0.0.1:3000
    - IPChandler/IPChandler -nodeID=2 -address=127.0.0.1:5002 -peers=127.0.0.1:5001,127.0.0.1:5003 -class=person -BS=127.0.0.1:3000
    - IPChandler/IPChandler -nodeID=3 -address=127.0.0.1:5003 -peers=127.0.0.1:5001,127.0.0.1:5002 -class=person -BS=127.0.0.1:3000

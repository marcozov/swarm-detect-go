# Bandwidth efficient object recognition for drone swarms

### DESCRIPTION:
The aim of this project is to improve object detection accuracy by taking advantage of multiple drones' viewpoints.

### RESULTS:
Our system is able to improve detection accuracy by eliminating some false positives and false negatives. Detailed comparisons, in terms of precision and recall, between our multi-host system and a single-host system are shown in the report.

### STRUCTURE:
    - IPCHandler/ : folder containing the inter process communication handler (for getting data from the python process through UNIX socket)
    - client/ : folder containing test clients scripts
    - structures/ : folder containing structures and functions that handle the communication with other hosts

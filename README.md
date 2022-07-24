<!-- ABOUT THE PROJECT -->
## About the Hijackdetector

This project implements a (near) real-time detection of potential BGP Hijacks.
In this project a potential BGP Hijack is defined as two BGP Update messages that announce subnets with at least on IP address in common, but with different Origin ASes.
To achieve this, we connect to the RIPE Routing Information System (https://ris-live.ripe.net/) via a livestream.
For each Announcement and Withdrawal messages relevant information (subnet, Origin AS, etc.) will be stored in a PATRICIA trie. 
During the traversal of the trie potential BGP Hijacks will be detected.

### Getting Started
1. Clone the repository: ````git clone https://github.com/RBReif/bgp-hijack-detection.git````
2. [OPTIONAL] Build the binary locally: ``go build -o hijackdetecter``
3. Execute the binary: ``./hijackdetecter``
4. [ALTERNATIVE] As an alternative to step 2. and 3. you can run (without building) the application: ``go run *.go``

### Analysing Memory and CPU consumption
Hijack Detector offers support to keep track of memory and CPU consumption of the Hijackdetector. 
The corresponding profile files are stored in the "output" folder (which will be automatically created if it does not yet exist).
You can enter the interactive analysis mode with ``go tool pprof PROFILENAME``. Per default the names of the profiles are cp (for the CPU profile) and mp (for the memory profile).
Individual names can be defined via flags. 

### Current Status
Currently, Hijackdetector only establishes the live stream to RIPE RIS and filters based for UPDATE messages. These are displayed. 
The next step is to handle and store update messages respectively. Afterwards fitting analysis functionality will be introduced. 

## Acknowledgements

### Previous Work at TUM
This project is based on a previous work by myself during my Master studies in Computer Science at the Technical University of Munich.  
I thank Prof. Carle, Mr. Sattler, and Mr. Zirngibl from the Chair of Network Architectures and Services for their support as my supervisor and advisors.

### Project Initiative by BND
This project is developed for the "Summer of Code" initiative of the Bundesnachrichtendienst (Foreign Intelligence Service of Germany) (https://www.bnd.bund.de/DE/Karriere/SummerOfCode/SummerOfCode_node.html). 

<!-- Contact -->
## Contact
Roland Reif - reifr@in.tum.de - www.linkedin.com/in/roland-reif/

Project Link: https://github.com/RBReif/bgp-hijack-detection


<!-- ABOUT THE PROJECT -->
## About the Hijackdetector

This project implements a (near) real-time detection of potential BGP Hijacks.
In this project a potential BGP Hijack is defined as two BGP Update messages that announce subnets with at least on IP address in common, but with different Origin ASes.
To achieve this, we connect to the RIPE Routing Information System (https://ris-live.ripe.net/) via a livestream.
Additionally, we can parse RIB and updates files from Routeviews.org.
For each Announcement and Withdrawal message relevant information (subnet, Origin AS, etc.) will be stored in a PATRICIA trie. 
During the traversal of the trie potential BGP Hijacks will be detected.

### Getting Started
1. Clone the repository: ````git clone https://github.com/RBReif/bgp-hijack-detection.git````
2. [OPTIONAL] Build the binary locally: ``go build -o hijackdetecter``
3. Execute the binary: ``./hijackdetecter``
4. [ALTERNATIVE] As an alternative to step 2. and 3. you can run (without building) the application: ``go run *.go``

### Usage examples
1. Only for Live Analysis without reading and parsing RIBs or updates files first. This is the default: ``go run *.go``
2. Only for Analysis of all updates files (e.g. "updates.20220826.1815.bz2") which followed a specific RIB (e.g. "rib.20220826.1800.bz2"), with all files being stored in one subdirectory (e.g. "input"): ``go run *.go -input="input" -rib="rib.20220826.1800.bz2" -live=false ``
3. As a combination of both, where first a specific RIB is parsed, then updates files are parsed and analysed for conflicts and then the analysis continues with the livefeed ``go run *.go -input="input" -rib="rib.20220826.1800.bz2"  ``
4. If you want the Live Analysis to stop at a specific point in time, you can do so with the flag -endLive: ``-endLive: 20220828.2000``

### RIB and updates files from Routeview.org
Hijackdetector supports .bz2, gz, and normal file formats. The naming convention is YYYYMMDD.HHMM. 
To sum it up the following file names are acceptable: [rib|updates].YYYYMMDD.HHMM[bz2|gz|]
The RIB and the updates files need to be in the same subdirectory.
If no RIB is specified, Hijackdetector will automatically use the newest RIB and all following updates files.
If no RIB is present in the specified subdirectory, all included updates files are parsed and analysed for conflicts.

### Analysing Memory and CPU consumption
Hijack Detector offers support to keep track of memory and CPU consumption of the Hijackdetector. 
You can enter the interactive analysis mode with ``go tool pprof PROFILENAME``. Per default the names of the profiles are cp (for the CPU profile) and mp (for the memory profile).
Individual names can be defined via flags (``-cpuprofile="myname"``, ``-memprofile="myothername"``).

### Further flags
* with ``-verbose=true`` you can print out more information to standard output
* with ``-risclient="your usecase"`` you can specify for what purposes you connect to RIPE RIS
* with ``-buffer=32000`` you can specify the maximum number of RIS messages to queue locally (in the exmaple to 32000)
* with ``-stream="your URL"`` you can specify a different input livestream source, if needed

### Current Status
Currently, Hijackdetector already offers the following features:
* Livestream to RIPE RIS
* Parsing of RIB files
* Parsing of updates. files
* Storing and inserting BGP update messages (both Announcements and Withdrawals) from above mentioned sources in a PATRICIA trie
* Conflict detection via traversing the trie
* Peer Awareness (a Withdrawal message is only eliminating Announcement messages from the same peer)

TODO: 
* writing found conflicts in e.g. a JSON file
* analysing found conflicts 

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


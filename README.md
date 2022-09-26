<!-- ABOUT THE PROJECT -->
## About the Hijackdetector

This project implements a (near) real-time detection of potential BGP Hijacks.
In this project a potential BGP Hijack is defined as two BGP Update messages that announce subnets with at least on IP address in common, but with different Origin ASes.
To achieve this, we connect to the RIPE Routing Information System (https://ris-live.ripe.net/) via a livestream.
Additionally, we can parse RIB and updates files from Routeviews.org.
For each ANNOUNCEMENT and WITHDRAWAL message relevant information (subnet, Origin AS, etc.) will be stored in a PATRICIA trie. 
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
4. If you want the Live Analysis to stop at a specific point in time, you can do so with the flag -endLive: ``-endLive=20220828.2000``
5. If you want an alert system for printing out conflicts involving any one of a list of predefined prefixes use ``go run *.go -prefixesfile="input/mySpecialPrefixes"``

### RIB and updates files from Routeview.org
Hijackdetector supports .bz2, gz, and normal file formats. The naming convention is YYYYMMDD.HHMM. 
To sum it up the following file names are acceptable: [rib|updates].YYYYMMDD.HHMM[bz2|gz|]
The RIB and the updates files need to be in the same subdirectory.
If no RIB is specified, Hijackdetector will automatically use the newest RIB and all following updates files.
If no RIB is present in the specified subdirectory, all included updates files are parsed and analysed for conflicts.

### Special monitoring of predefined prefixes
With the ``-prefixesfile`` flag, it is possible to provide a list of prefixes where any conflict involving one announcement for such a prefix (or a more specific prefix) gets directly printed out to standard output.
The file is expected to contain one IPv4 prefix per line with the network address followed by a slash (/) and the subnet length.

### Analysis of found conflicts
All found conflicts will be written to a file.
With ``-conflictsfile=yourFileName.json`` you can set the name and location of the file. (By default this file is called conflicts.json and located in the output directory).
Every conflict consists of exactly one "referenceAnnouncement" (the update message, which triggered a conflict) and one or multiple "conflicts".
Both the "referenceAnnouncement" and all "conflicts" are of the same type and consist of "subnet", "origin", "timestamp", and "aspath" (where "origin" is the last AS in "aspath).

### Analysis of participating ASes
All ASes which appear as "origin" in a conflicts will be written to a .csv file alongside further quantitative and qualitative attributes.
With ``-originsfile=yourFileName.csv`` you can set the name and location of the file. (By default this file is called origins.csv and located in the output directory).
The format is: asn, total, lessSpecificOrigin, sameSubnet, moreSpecificOrigin, legit, registry, country, ispName.
* asn: the number of the autonomous systems, appearing as an origin-AS
* total: the total number of appearances as an origin in a conflict. It equals lessSpecificOrigin+sameSubnet+moreSpecificOrigin
* lessSpecificOrigin: the number of times in which the AS appeared as origin of the ANNOUNCEMENT for the less specific subnet. (A high number might indicate that this AS has been victim to a lot of BGP Hijacks.)
* sameSubnet: the number of times in which the AS appeared as origin of an ANNOUNCEMENT in conflict with another ANNOUNCEMENT of exactly the same subnet
* moreSpecificOrigin: the number of times in which the AS appeared as origin of the ANNOUNCEMENT for the more specific subnet. (A high number might indicate that this AS has been an attacker of a lot of BGP Hijacks.)
* legit: the number of times in which one of the origins of two conflicting ANNOUNCEMENTs is in the AS-path of the other ANNOUNCEMENT, and hence indicating a topological (probably legit) relation
* registry: the registry responsible for the AS
* country: the country where the AS is positioned
* ispName: the name of the corresponding Internet Service Provider. Note, that often the Country Code is included also in the end of the ISP name

A short summary of the origin ASes for the most often appearing ASes is also printed after each run of Hijack Detector to standard output.
For simplicity reasons "lessSpecificOrigin" is written as "victim", and "moreSpecificOrigin" is written as "attacker" in the printed overview.

### Analysing Memory and CPU consumption
Hijack Detector offers support to keep track of memory and CPU consumption of the Hijackdetector. 
You can enter the interactive analysis mode with ``go tool pprof PROFILENAME``. Per default the names of the profiles are cp (for the CPU profile) and mp (for the memory profile).
By default they will be stored in the output directory.
Individual names can be defined via flags (``-cpuprofile="myname"``, ``-memprofile="myothername"``).

### Further flags
* with ``-verbose=true`` you can print out more information to standard output
* with ``-risclient="your usecase"`` you can specify for what purposes you connect to RIPE RIS
* with ``-buffer=32000`` you can specify the maximum number of RIS messages to queue locally (in the exmaple to 32000)
* with ``-stream="your livestream source URL"`` you can specify a different input livestream source, if needed
* with ``-ribconflicts=true`` you can already find conflicts in a specified RIB file itself


### Stop the program
With SIGTERM (e.g. Ctrl+C) you can gracefully end prgoram execution and print out some stats. 
Livemode can also be ended the program with the ``-endlive`` flag

### Current Status
Currently, Hijackdetector already offers the following features:
* Livestream to RIPE RIS
* Parsing of RIB files
* Parsing of updates. files
* Storing and inserting BGP update messages (both ANNOUNCEMENTS and WITHDRAWALS) from above mentioned sources in a PATRICIA trie
* Conflict detection via traversing the trie
* Peer Awareness (a WITHDRAWAL message is only eliminating ANNOUNCEMENT messages from the same peer)
* Writing found conflicts in a .JSON file
* Analysing ASes which were potentially a victim or an attacker during BGP hijacks (based on frequency, topological relations and retrieving background information)
* Additional, special monitoring of conflicts inside specified prefixes is now also implemented

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


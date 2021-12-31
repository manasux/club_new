## IMPORTANT NOTES

The current club is still under intense development. System structure might change at any time. Please only develop on the current existing club Gateway Interface (AGI) JavaScript Interface or standard HTML webapps with ao_module.js endpoints.

## Features

### User Interface

- Web Desktop Interface (Better than Synology DSM)
- Ubuntu remix Windows style startup menu and task bars
- Clean and easy to use File Manager (Support drag drop, upload etc)
- Simplistic System Setting Menu
- No-bull-shit module naming scheme

### Networking

- FTP Server
- WebDAV Server
- UPnP Port Forwarding
- Samba (Supported via 3rd party sub-services)
- WiFi Management (Support wpa_supplicant for Rpi or nmcli for Armbian)

### File / Disk Management

- Mount / Format Disk Utilities (support NTFS, EXT4 and more!)
- Virtual File System Architecture
- File Sharing (Similar to Google Drive)
- Basic File Operations with Real-time Progress (Copy / Cut / Paste / New File or Folder etc)

### Others

- Require as little as 512MB system memory and 8GB system storage
- Base on one of the most stable Linux distro - Debian
- Support for Desktop, Laptop (touchpad) and Mobile screen sizes

## Screenshots

![Image](img/screenshots/1.png?raw=true)
![Image](img/screenshots/2.png?raw=true)
![Image](img/screenshots/3.png?raw=true)
![Image](img/screenshots/4.png?raw=true)
![Image](img/screenshots/5.png?raw=true)
![Image](img/screenshots/6.png?raw=true)

## Start the club Platform

### Supported Startup Parameters

The following startup parameters are supported (v1.113)

```
-allow_autologin
    	Allow RESTFUL login redirection that allow machines like billboards to login to the system on boot (default true)
  -allow_cluster
    	Enable cluster operations within LAN. Require allow_mdns=true flag (default true)
  -allow_iot
    	Enable IoT related APIs and scanner. Require MDNS enabled (default true)
  -allow_mdns
    	Enable MDNS service. Allow device to be scanned by nearby AOZ Hosts (default true)
  -allow_pkg_install
    	Allow the system to install package using Advanced Package Tool (aka apt or apt-get) (default true)
  -allow_ssdp
    	Enable SSDP service, disable this if you do not want your device to be scanned by Windows's Network Neighborhood Page (default true)
  -allow_upnp
    	Enable uPNP service, recommended for host under NAT router
  -beta_scan
    	Allow compatibility to club Online Beta Clusters
  -cert string
    	TLS certificate file (.crt) (default "localhost.crt")
  -console
    	Enable the debugging console.
  -demo_mode
    	Run the system in demo mode. All directories and database are read only.
  -dir_list
    	Enable directory listing (default true)
  -disable_http
    	Disable HTTP server, require tls=true
  -disable_ip_resolver
    	Disable IP resolving if the system is running under reverse proxy environment
  -disable_subservice
    	Disable subservices completely
  -enable_hwman
    	Enable hardware management functions in system (default true)
  -gzip
    	Enable gzip compress on file server (default true)
  -homepage
    	Enable user homepage. Accessible via /www/{username}/ (default true)
  -hostname string
    	Default name for this host (default "My club")
  -iobuf int
    	Amount of buffer memory for IO operations (default 1024)
  -key string
    	TLS key file (.key) (default "localhost.key")
  -max_upload_size int
    	Maxmium upload size in MB. Must not exceed the available ram on your system (default 8192)
  -ntt int
    	Nightly tasks execution time. Default 3 = 3 am in the morning (default 3)
  -port int
    	Listening port for HTTP server (default 8080)
  -public_reg
    	Enable public register interface for account creation
  -root string
    	User root directories (default "./files/")
  -session_key string
    	Session key, must be 16, 24 or 32 bytes long (AES-128, AES-192 or AES-256). Leave empty for auto generated.
  -storage_config string
    	File location of the storage config file (default "./system/storage.json")
  -tls
    	Enable TLS on HTTP serving (HTTPS Mode)
  -tls_port int
    	Listening port for HTTPS server (default 8443)
  -tmp string
    	Temporary storage, can be access via tmp:/. A tmp/ folder will be created in this path. Recommend fast storage devices like SSD (default "./")
  -tmp_time int
    	Time before tmp file will be deleted in seconds. Default 86400 seconds = 24 hours (default 86400)
  -upload_async
    	Enable file upload buffering to run in async mode (Faster upload, require RAM >= 8GB)
  -upload_buf int
    	Upload buffer memory in MB. Any file larger than this size will be buffered to disk (slower). (default 25)
  -uuid string
    	System UUID for clustering and distributed computing. Only need to config once for first time startup. Leave empty for auto generation.
  -version
    	Show system build version
  -wlan_interface_name string
    	The default wireless interface for connecting to an AP (default "wlan0")
  -wpa_supplicant_config string
    	Path for the wpa_supplicant config (default "/etc/wpa_supplicant/wpa_supplicant.conf")
```

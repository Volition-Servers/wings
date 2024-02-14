[![Logo Image](https://cdn.pterodactyl.io/logos/new/pterodactyl_logo.png)](https://pterodactyl.io)

![Discord](https://img.shields.io/discord/122900397965705216?label=Discord&logo=Discord&logoColor=white)
![GitHub Releases](https://img.shields.io/github/downloads/pterodactyl/wings/latest/total)
[![Go Report Card](https://goreportcard.com/badge/github.com/pterodactyl/wings)](https://goreportcard.com/report/github.com/pterodactyl/wings)

# Pterodactyl Wings

Wings is Pterodactyl's server control plane, built for the rapidly changing gaming industry and designed to be
highly performant and secure. Wings provides an HTTP API allowing you to interface directly with running server
instances, fetch server logs, generate backups, and control all aspects of the server lifecycle.

In addition, Wings ships with a built-in SFTP server allowing your system to remain free of Pterodactyl specific
dependencies, and allowing users to authenticate with the same credentials they would normally use to access the Panel.

## Sponsors

I would like to extend my sincere thanks to the following sponsors for helping find Pterodactyl's development.
[Interested in becoming a sponsor?](https://github.com/sponsors/matthewpi)

| Company                                                   | About                                                                                                                                                                                                 |
|-----------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [**Aussie Server Hosts**](https://aussieserverhosts.com/) | No frills Australian Owned and operated High Performance Server hosting for some of the most demanding games serving Australia and New Zealand.                                                       |
| [**BisectHosting**](https://www.bisecthosting.com/)       | BisectHosting provides Minecraft, Valheim and other server hosting services with the highest reliability and lightning fast support since 2012.                                                       |
| [**MineStrator**](https://minestrator.com/)               | Looking for the most highend French hosting company for your minecraft server? More than 24,000 members on our discord trust us. Give us a try!                                                       |
| [**VibeGAMES**](https://vibegames.net/)                   | VibeGAMES is a game server provider that specializes in DDOS protection for the games we offer. We have multiple locations in the US, Brazil, France, Germany, Singapore, Australia and South Africa. |

## Documentation

* [Panel Documentation](https://pterodactyl.io/panel/1.0/getting_started.html)
* [Wings Documentation](https://pterodactyl.io/wings/1.0/installing.html)
* [Community Guides](https://pterodactyl.io/community/about.html)
* Or, get additional help [via Discord](https://discord.gg/pterodactyl)

## Reporting Issues

Please use the [pterodactyl/panel](https://github.com/pterodactyl/panel) repository to report any issues or make
feature requests for Wings. In addition, the [security policy](https://github.com/pterodactyl/panel/security/policy) listed
within that repository also applies to Wings.

## Install Instructions

1, Make a directory for wings source code and go to it: mkdir /srv/wings && cd /srv/wings

2, Download the latest release with wget or manually from github: https://github.com/pterodactyl/wings
   a, Select the latest release and download, for example: wget https://github.com/pterodactyl/wings/archive/release/v1.3.0.zip
   b, Select the latest release and download with the Code -> Download ZIP Button and upload manually to your server

3, Unzip it: unzip  v1.3.0.zip

4, Paste WingsFiles to your wings folder

5, 
a, Open /router/router.go
   
b, Paste this line to under the server.POST("/ws/deny", postServerDenyWSTokens) line
      
      server.POST("/versions/list", getVersions)
      server.POST("/versions/switch", switchVersion)

6, 
Open /router/router.go

Paste this line to under the files.POST("/chmod", postServerChmodFile) line

      files.POST("/search/smart", smartSearch)

7, Install go: https://golang.org/doc/install

8, Build the new wings (if you created to other folder, change it - or if your wings is other folder, change it): 
   - cd /srv/wings/ && go build -o /usr/local/bin/wings && chmod +x /usr/local/bin/wings

9, Restart wings: service wings restart

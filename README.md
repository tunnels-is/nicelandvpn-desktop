[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
<!-- [![GNU License][license-shield]][license-url] -->
<!-- [![LinkedIn][linkedin-shield]][linkedin-url] -->

# Niceland VPN 
This repository contains the Niceland VPN desktop application. </br>
Anyone is welcome to contribute, just fork and make a pull request, or you can ask to become a collaborator. 

# GPU Acceleration
If you are running NicelandVPN on Linux, you might need to disable GPU acceleration. That can be done by running Niclenad with the -disableGPU flag
```
$ ./NicelandVPN -disableGPU
```

# Building from source
 - [install nodejs and npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)
 - install npm packages for the frontend
 ```
$ cd [PROJECT_ROOT]/frontend
$ npm install
 ```
 - [install golang](https://go.dev/doc/install)
 - [install wails.io](https://wails.io/docs/gettingstarted/installation)
 - build the app
 ```
$ cd [PROJECT_ROOT]
 wails build -webview2 embed
 ```
 - the final build product can be found inside the build/bin folder

# Setup up for development
 - [install nodejs and npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm)
 - install npm packages for the frontend
 ```
$ cd [PROJECT_ROOT]/frontend
$ npm install
 ```
 - [install Golang](https://go.dev/doc/install)
 - [install Wails.io](https://wails.io/docs/gettingstarted/installation)
 - start the Wails development server
 - NOTE: before starting the server we recommend setting ENABLE_INTERFACE to false in the main.go file. This will prevent Wails from initializing the default tunnel tap interface every time it reloads due to changes in files.
 ```
$ cd [PROJECT_ROOT]
 wails dev
 ```

# General development guide
 - Indentation: 1 Tab.
 - Naming should follow a Camel Case convention. Some exceptions may apply.

## Folders
 - **build/bin:** final build products
 - **build/windows-custom:** Windows build staging grounds, build scripts, sign scripts, etc.
 - **build/darwin-custom:** MacOs build staging grounds, sign scripts, .app files, etc.
 - **core:** the code for the VPN client 
 - **frontend/src:** The GUI source files
 - **frontend/Wailsjs:** Wails.io generated javascript files
 - **launcher:** MacOS-specific launching mechanism
 - **parking:** code examples that might be interesting to look at/use later. 

## Notes on the Golang core 
 - Yes, there are race conditions. Most of them are on purpose. If you want to fix race conditions, then please create an issue first so we can discuss it. 
 - Avoid using Locks or sync maps whenever possible. 
 - Sacrificing memory for performance is generally accepted, however, we do not want to use up too much memory. Changes that implement this trade-off need to be evaluated.
 - Niceland has a few async routines that manage the state, and we would like to keep these routines to a minimum. If you want to implement something that needs to happen on an interval, it should preferably be implemented in the "StateMaintenance" routine.

## Using ReactJS
 - Custom functions and methods should start with an uppercase letter. 

 - Import are sorted by type:
    1. ReactJS imports
    2. Modules from NPM
    3. Wails / Golang bindings
    4. Internal components and libraries

</br>

 - We want to keep the complexity of this project to a minimum. Any ReactJS functionality that is not generally accepted will need to be reviewed on a per-case basis.

 - ReactJS functionality that is generally accepted:
    - useState
    - useEffect
    - useNavigate

# Pull Requests
 - No copy/paste from machine learning tools.

Pull requests should have a short descriptive title and a long description when needed, listing everything that has been modified or added, but more importantly why it was modified or added. 


# AdBlock sources
 - https://github.com/badmojr/1Hosts

[forks-shield]: https://img.shields.io/github/forks/tunnels-is/nicelandvpn-desktop?style=for-the-badge&logo=github
[forks-url]: https://github.com/tunnels-is/nicelandvpn-desktop/network/members
[stars-shield]: https://img.shields.io/github/stars/tunnels-is/nicelandvpn-desktop?style=for-the-badge&logo=github
[stars-url]: https://github.com/tunnels-is/nicelandvpn-desktop/stargazers
[issues-shield]: https://img.shields.io/github/issues/tunnels-is/nicelandvpn-desktop?style=for-the-badge&logo=github
[issues-url]: https://github.com/tunnels-is/nicelandvpn-desktop/issues
<!-- [license-shield]: https://img.shields.io/github/license/umutsevdi/Logic-Circuit-Simulator.svg?style=for-the-badge -->
<!-- [license-url]: https://github.com/umutsevdi/Logic-Circuit-Simulator/blob/main/LICENSE -->
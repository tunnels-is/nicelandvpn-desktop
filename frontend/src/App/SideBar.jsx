import { useNavigate, useLocation } from "react-router-dom";
import React, { useEffect, useState } from "react";

import { ActivityLogIcon, BarChartIcon, ChatBubbleIcon, DesktopIcon, EnterIcon, ExitIcon, ExternalLinkIcon, FileTextIcon, GearIcon, GlobeIcon, HamburgerMenuIcon, LinkBreak2Icon, MobileIcon, Share1Icon, } from '@radix-ui/react-icons'
import Loader from "react-spinners/PacmanLoader";

import { OpenURL } from "../../wailsjs/go/main/App";

import STORE from "../store";

const SideBar = (props) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [menuTab, setMenuTab] = useState(1)

  let { pathname } = location
  let sp = pathname.split("/")

  const clickHandler = (path) => {
    console.log("navigating to:", path)
    navigate(path)
  }

  const HandleLogout = async () => {
    props.toggleLoading({ logTag: "disconnect", tag: "LOGOUT", show: true, msg: "Disconnecting and logging out...", includeLogs: true })
    await props.disconnectFromVPN()
    STORE.CleanupOnLogout()
    navigate("/login")
  }


  let user = STORE.GetUser()
  let hasSub = false
  let needsUpdate = false

  if (user) {
    if (user.Version !== props.state.Version) {
      needsUpdate = true
    }

    if (user.SubLevel === 666) {
      hasSub = false
    } else if (user.SubLevel > 0) {
      hasSub = true
    }
  }

  const RenderPopovers = () => {
    return (
      <div className="popover-container">
        {needsUpdate &&
          <div className="popover">
            A new version is available <br />

            <a onClick={() => OpenURL("https://nicelandvpn.is")}>Please visit the website to download the latest version</a>
          </div>
        }

        {(!hasSub && user) &&
          <div className="popover">
            No active subscription <br />
            <a onClick={() => OpenURL("https://www.nicelandvpn.is/#/pricing")} >Visit the website to subscribe</a>
          </div>
        }
      </div >
    )
  }

  if (menuTab === 2) {
    if (!props.state) {
      return (
        <div className="stats-bar">
          <div className="back" onClick={() => setMenuTab(1)}> {`<`} Menu</div>
          <div className="title">Loading state ...</div>
        </div>
      )
    }

    return (
      <div className="stats-bar">
        {RenderPopovers()}
        <div className="back" onClick={() => setMenuTab(1)}>
          <HamburgerMenuIcon className="icon" height={20} width={20}></HamburgerMenuIcon>
          <div className="text">
            {` Back To Menu`}
          </div>
        </div>
        <div className="title">
          <DesktopIcon className="icon" height={20} width={20}></DesktopIcon>
          <div className="text">
            App State
          </div>
        </div>

        <div className="stats-item">
          <div className="label">VPN List Update</div>
          <div className="value">{props.state.SecondsUntilAccessPointUpdate}</div>
        </div>
        <div className="stats-item">
          <div className="label">Ready To Connect</div>
          <div className="value">{props.state.ClientReady + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Version</div>
          <div className="value">{props.state.Version + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">VPN Tunnel ready</div>
          <div className="value">{props.state.TunnelInitialized + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Launched As Admin</div>
          <div className="value">{props.state.IsAdmin + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Config Loaded</div>
          <div className="value">{props.state.ConfigInitialized + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Base Folder Created</div>
          <div className="value">{props.state.BaseFolderInitialized + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Log File Created</div>
          <div className="value">{props.state.LogFileInitialized + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Buffer Error</div>
          <div className="value">{props.state.BufferError + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Launch Error</div>
          <div className="value">{props.state.ClientStartupError + ""}</div>
        </div>

        {props.state?.DefaultInterface &&
          <>
            <div className="title">
              <LinkBreak2Icon className="icon" height={20} width={20}></LinkBreak2Icon>
              <div className="text">
                Interface
              </div>
            </div>
            <div className="stats-item">
              <div className="label">Name</div>
              <div className="value">{props.state.DefaultInterface.IFName}</div>
            </div>
            <div className="stats-item">
              <div className="label">IPv6 Enabled</div>
              <div className="value">{props.state.DefaultInterface.IPV6Enabled + ""}</div>
            </div>
            <div className="stats-item">
              <div className="label">Gateway</div>
              <div className="value">{props.state.DefaultInterface.DefaultRouter}</div>
            </div>
            {props.state?.DefaultInterface.AutoDNS === true &&
              <div className="stats-item">
                <div className="label">DNS from DHCP</div>
                <div className="value">{props.state.DefaultInterface.AutoDNS + ""}</div>
              </div>
            }
            {(props.state.DefaultInterface.DNS2 === "" && props.state.DefaultInterface.DNS1 !== "") &&
              <div className="stats-item">
                <div className="label">DNS</div>
                <div className="value">{props.state.DefaultInterface.DNS1}</div>
              </div>
            }

            {props.state.DefaultInterface.DNS2 !== "" &&
              <div className="stats-item">
                <div className="label">DNS</div>
                <div className="value">{props.state.DefaultInterface.DNS1}{" - "}{props.state.DefaultInterface.DNS2}</div>
              </div>
            }
          </>
        }

        <div className="title">
          <Share1Icon className="icon" height={20} width={20}></Share1Icon>
          <div className="text">
            Connection
          </div>
        </div>

        {props.state?.ActiveRouter &&
          <>
            <div className="stats-item">
              <div className="label">Entry Router</div>
              <div className="value">{props.state.ActiveRouter.Tag}</div>
            </div>
            <div className="stats-item">
              <div className="label">MS / QoS</div>
              <div className="value">{props.state.ActiveRouter.MS}{" / "}{props.state.ActiveRouter.Score} </div>
            </div>
          </>
        }

        {props.state?.ActiveAccessPoint &&
          <>
            <div className="stats-item">
              <div className="label">Exit Router</div>
              <div className="value">{props.state.ActiveAccessPoint.Router.Tag}</div>
            </div>
            <div className="stats-item">
              <div className="label">MS / QoS</div>
              <div className="value">{props.state.ActiveAccessPoint.Router.MS}{" / "}{props.state.ActiveAccessPoint.Router.Score}</div>
            </div>
            <div className="stats-item">
              <div className="label">VPN</div>
              <div className="value">{props.state.ActiveAccessPoint.Tag}</div>
            </div>
            {props.state?.Connected &&
              <div className="stats-item">
                <div className="label">Duration</div>
                <div className="value">{props.state.ConnectedTimer + ""}</div>
              </div>
            }
            <div className="stats-item">
              <div className="label">Last Ping</div>
              <div className="value">{props.state.SecondsSincePingFromRouter}</div>
            </div>
          </>
        }

        <div className="title">
          <ActivityLogIcon className="icon" height={20} width={20}></ActivityLogIcon>
          <div className="text">
            Network Stats
          </div>
        </div>
        <div className="stats-item">
          <div className="label">Connected</div>
          <div className="value">{props.state.Connected + ""}</div>
        </div>
        <div className="stats-item">
          <div className="label">Download</div>
          <div className="value">{props.state.DMbpsString}</div>
        </div>
        <div className="stats-item">
          <div className="label">Packets</div>
          <div className="value">{props.state.IngressPackets}</div>
        </div>
        <div className="stats-item">
          <div className="label">Upload</div>
          <div className="value">{props.state.UMbpsString}</div>
        </div>
        <div className="stats-item">
          <div className="label">Packets</div>
          <div className="value">{props.state.EgressPackets}</div>
        </div>

      </div>
    )

  } else if (menuTab === 1) {
    return (
      <div className="side-bar">
        {RenderPopovers()}
        {props.loading &&
          <div className="loader-wrapper">
            <Loader
              className="loader"
              size={56}
              color={"#FF922D"}
            ></Loader>
          </div>
        }
        {!props.loading &&
          <div className={`logo`}  >
          </div>
        }

        <div className="menu-items">
          {!user &&
            <div className={`menu-link  ${sp[1] == "login" ? "menu-active" : ""}`}
              onClick={() => clickHandler("/login")} >
              <EnterIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></EnterIcon>
              <div className="menu-text login">
                Login
              </div>
            </div>

          }
          {props.state?.ActiveAccessPoint &&
            <div className={`menu-link`}
              onClick={() => props.disconnectFromVPN()} >
              <ExternalLinkIcon width={30} height={30} color={"#d00707"} className="menu-list-icon"></ExternalLinkIcon>
              <div className="menu-text disconnect" >
                Disconnect
              </div>
            </div>
          }

          {user &&
            <div className={`menu-link  ${sp[1] == "" ? "menu-active" : ""}`}
              onClick={() => clickHandler("/")} >
              <GlobeIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></GlobeIcon>
              <div className="menu-text vpns" >
                VPNs
              </div>
            </div>
          }

          {props.advancedMode &&

            <>
              <div className={`menu-link  ${sp[1] == "routers" ? "menu-active" : ""}`}
                onClick={() => clickHandler("/routers")} >
                <MobileIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></MobileIcon>
                <div className="menu-text routers">
                  Routers
                </div>
              </div>

              <div className={`menu-link ${sp[1] == "stats" ? "menu-active" : ""}`}
                onClick={() => setMenuTab(2)} >
                <BarChartIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></BarChartIcon>
                <div className="menu-text stats">
                  Stats
                </div>
              </div>

              <div className={`menu-link ${sp[1] == "logs" ? "menu-active" : ""}`}
                onClick={() => clickHandler("/logs")} >
                <FileTextIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></FileTextIcon>
                <div className="menu-text logs">
                  Logs
                </div>
              </div>
            </>
          }
          <div className={`menu-link ${sp[1] == "settings" ? "menu-active" : ""}`}
            onClick={() => clickHandler("/settings")} >
            <GearIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></GearIcon>
            <div className="menu-text settings">
              Settings
            </div>
          </div>
          <div className={`menu-link ${sp[1] == "support" ? "menu-active" : ""}`}
            onClick={() => clickHandler("/support")} >
            <ChatBubbleIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></ChatBubbleIcon>
            <div className="menu-text help">
              Help
            </div>
          </div>

          {user &&
            <div className={`menu-link`}
              onClick={() => HandleLogout()} >
              <ExitIcon width={30} height={30} color={"#d00707"} className="menu-list-icon"></ExitIcon>
              <div className="menu-text logout">
                Logout
              </div>
            </div>
          }


        </div>
      </div >
    )
  }

}

export default SideBar;
import { useNavigate, useLocation } from "react-router-dom";
import React, { useEffect, useState } from "react";

import { ActivityLogIcon, BarChartIcon, ChatBubbleIcon, DesktopIcon, EnterIcon, ExitIcon, ExternalLinkIcon, FileTextIcon, GearIcon, GlobeIcon, HamburgerMenuIcon, LinkBreak2Icon, MobileIcon, Share1Icon, } from '@radix-ui/react-icons'
import Loader from "react-spinners/PacmanLoader";
import toast from 'react-hot-toast';

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

  const ConfirmLogout = () => {

    toast.error((t) =>
    (
      <div className="exit-confirm">
        <div className="text">
          Are you sure you want to
        </div>
        <div className="text-big">
          Logout
        </div>
        <button className="exit" onClick={() => toast.dismiss(t.id)}>Cancel</button>
        <button className="cancel" onClick={() => {
          toast.dismiss(t.id)
          HandleLogout()
        }
        }>Confirm</button>
      </div>

    ), { id: "logout", duration: 999999 }
    )
  }

  const ConfirmDisconnect = () => {

    toast.error((t) =>
    (
      <div className="exit-confirm">
        <div className="text">
          Are you sure you want to
        </div>
        <div className="text-big">
          Disconnect
        </div>
        <button className="exit" onClick={() => toast.dismiss(t.id)}>Cancel</button>
        <button className="cancel" onClick={() => {
          toast.dismiss(t.id)
          props.disconnectFromVPN()
        }
        }>Confirm</button>
      </div>

    ), { id: "logout", duration: 999999 }
    )
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

            <div onClick={() => OpenURL("https://nicelandvpn.is")}>Please visit the website to download the latest version</div>
          </div>
        }

        {(!hasSub && user) &&
          <div className="popover">
            No active subscription <br />
            <div onClick={() => OpenURL("https://www.nicelandvpn.is/#/pricing")} >Visit the website to subscribe</div>
          </div>
        }
      </div >
    )
  }

  if (!props.state) {
    return (
      <div className="stats-bar">
        <div className="back" onClick={() => setMenuTab(1)}> {`<`} Menu</div>
        <div className="title">Loading state ...</div>
      </div>
    )
  }


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

            {!props.stats &&
              <div className={`menu-link ${sp[1] == "stats" ? "menu-active" : ""}`}
                onClick={() => props.setStats(true)} >
                <BarChartIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></BarChartIcon>
                <div className="menu-text stats">
                  Stats
                </div>
              </div>
            }

            {props.stats &&
              <div className={`menu-link ${sp[1] == "stats" ? "menu-active" : ""}`}
                onClick={() => props.setStats(false)} >
                <BarChartIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></BarChartIcon>
                <div className="menu-text stats">
                  Stats
                </div>
              </div>
            }


          </>
        }



        <div className={`menu-link ${sp[1] == "settings" ? "menu-active" : ""}`}
          onClick={() => clickHandler("/settings")} >
          <GearIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></GearIcon>
          <div className="menu-text settings">
            Settings
          </div>
        </div>

        <div className={`menu-link ${sp[1] == "logs" ? "menu-active" : ""}`}
          onClick={() => clickHandler("/logs")} >
          <FileTextIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></FileTextIcon>
          <div className="menu-text logs">
            Logs
          </div>
        </div>

        <div className={`menu-link ${sp[1] == "support" ? "menu-active" : ""}`}
          onClick={() => clickHandler("/support")} >
          <ChatBubbleIcon width={30} height={30} color={"#20C997"} className="menu-list-icon"></ChatBubbleIcon>
          <div className="menu-text help">
            Help
          </div>
        </div>

        {props.state?.ActiveAccessPoint &&
          <div className={`menu-link`}
            onClick={() => ConfirmDisconnect()} >
            <ExternalLinkIcon width={30} height={30} color={"#d00707"} className="menu-list-icon"></ExternalLinkIcon>
            <div className="menu-text disconnect" >
              Disconnect
            </div>
          </div>
        }


        {user &&
          <div className={`menu-link`}
            onClick={() => ConfirmLogout()} >
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

export default SideBar;
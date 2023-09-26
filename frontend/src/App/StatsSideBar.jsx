import React from "react";

import { ActivityLogIcon, DesktopIcon, HamburgerMenuIcon, LinkBreak2Icon, MobileIcon, Share1Icon, } from '@radix-ui/react-icons'


const StatsSideBar = (props) => {

  if (!props.state) {
    return (
      <div className="stats-sidebar">
        <div className="title">Loading state ...</div>
      </div>
    )
  }

  return (
    <div className="stats-sidebar">

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

}

export default StatsSideBar;
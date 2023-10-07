import { Navigate } from "react-router-dom";
import React, { useState } from "react";

import { SwitchRouter } from '../../wailsjs/go/main/Service';

import STORE from "../store";
import { DesktopIcon, MagnifyingGlassIcon, EnterIcon } from "@radix-ui/react-icons";

const Routers = (props) => {

	const [filter, setFilter] = useState("");

	const switchRouter = async (router) => {

		if (props.state?.Connected) {
			props.toggleError("Unable to change routers while connected")
			return
		}


		if (router.Tag === "") {
			props.toggleLoading({ tag: "ROUTERS", show: true, msg: "Switching to automatic router selection" })
		} else {
			props.toggleLoading({ tag: "ROUTERS", show: true, msg: "Switching to " + router.Tag })
		}

		SwitchRouter(router.Tag).then((x) => {
			if (x.Err) {
				props.toggleError(x.Err)
			} else {
				props.showSuccessToast("Router switch complete")
			}
		}).catch((error) => {
			console.dir(error)
			props.toggleError("Unknown error, please try again in a moment");
		})

		props.toggleLoading(undefined)
	}


	let routers = []

	if (props?.state?.Routers) {

		if (filter && filter !== "") {
			props.state.Routers.map(r => {

				let filterMatch = false
				if (r.Tag.includes(filter)) {
					filterMatch = true
				}

				if (filterMatch) {
					routers.push(r)
				}
			})

		} else {
			routers = props.state.Routers
		}

	}
	console.log("ROUTER COUNT:", routers.length)
	console.dir(routers)

	const RenderServer = (s, active) => {


		return (
			<div className="server" key={s.Tag} onClick={() => switchRouter(s)}>
				{/* <div className="connect"></div> */}
				{s.Tag &&
					<div className="item ip">
						{active &&

							<EnterIcon className="icon"></EnterIcon>
						}
						{s.Tag}
					</div>
				}
				{!s.Tag &&
					<div className="item ip">Unknown</div>
				}
				{s.Country !== "" &&
					<div className="item country-item x3">
						<img
							className="flag"
							src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + s.Country.toLowerCase() + ".svg"}
						// src={"/src/assets/images/flag/" + s.Country.toLowerCase() + ".svg"}
						/>
						<span className="name">
							{s.Country}
						</span>
					</div>
				}
				{s.Country === "" &&
					<div className="item country-item x3">
						<span className="name">
						...
						</span>
					</div>

				}
				<div className="item x3">{s.Score}</div>
				<div className="item x3">{s.MS === 999 ? "?" : s.MS}</div>
				<div className="item x3">{(s.AvailableSlots)}</div>
				<div className="item x3">{s.AvailableMbps / 1000} </div>
				<div className="item x3">{s.CPUP}</div>
				<div className="item x3">{s.RAMUsage}</div>
				<div className="item x3">{s.DiskUsage}</div>
			</div>
		)
	}

	let AR = props.state?.ActiveRouter

	return (
		<div className="router-wrapper"  >

			<div className="search-wrapper">
				<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
				<input type="text" className="search" onChange={(e) => setFilter(e.target.value)} placeholder="Search .."></input>
			</div>

			{props.state?.C?.ManualRouter &&
				<div className="automatic-button"
					onClick={() => switchRouter({ Tag: "" })} >Switch Back To Automatic Router Selection</div>
			}

			<div className="router-list">

				<div className="header">
					<div className="title ip">Tag
					</div>
					<div className="title x3">Country</div>
					<div className="title x3">QoS
						<span className="tooltiptext">{STORE.ROUTER_Tooltips[0]}</span>
					</div>
					<div className="title x3">MS
						<span className="tooltiptext">{STORE.ROUTER_Tooltips[1]}</span>
					</div>
					<div className="title x3">Slots
						<span className="tooltiptext">{STORE.ROUTER_Tooltips[2]}</span>
					</div>
					<div className="title x3">Gbps
						<span className="tooltiptext">{STORE.ROUTER_Tooltips[3]}</span>
					</div>
					<div className="title x3">CPU %
						<span className="tooltiptext">{STORE.ROUTER_Tooltips[4]}</span>
					</div>
					<div className="title x3">RAM %
						<span className="tooltiptext">{STORE.ROUTER_Tooltips[5]}</span>
					</div>
					<div className="title x3">DISK %
						<span className="tooltiptext">{STORE.ROUTER_Tooltips[6]}</span>
					</div>
				</div>


				{routers.map(r => {
					if (AR) {
						if (AR.Tag === r.Tag) {
							return RenderServer(r, true)
						} else {
							return RenderServer(r, false)
						}
					} else {
						return RenderServer(r, false)
					}
				})}

			</div>

		</div >
	);
}

export default Routers;

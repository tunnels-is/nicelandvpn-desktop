import { useNavigate, Navigate } from "react-router-dom";
import React, { useState } from "react";

import toast from 'react-hot-toast';
import Loader from "react-spinners/ScaleLoader";

import { Connect, Switch } from "../../wailsjs/go/main/Service";

import STORE from "../store";
import { DesktopIcon, EnterIcon, MagnifyingGlassIcon } from "@radix-ui/react-icons";


const Dashboard = (props) => {

	const [filter, setFilter] = useState("");
	const navigate = useNavigate();

	const updateFilter = (event) => {
		setFilter(event.target.value)
	}

	const LogOut = () => {
		props.toggleError("You have been logged out")
		STORE.Cache.Clear()
	}


	const ConfirmConnect = (a, ar) => {

		toast.success((t) => (
			<div className="exit-confirm">
				<div className="text">
					Your are connecting to
				</div>
				<div className="text-big">
					{a.Tag}
				</div>
				<button className="cancel" onClick={() => {
					toast.dismiss(t.id)
					ConnectToVPN(a, ar)
				}
				}>Connect</button>
				<button className="exit" onClick={() => toast.dismiss(t.id)}>Cancel</button>
			</div>
		), { id: "connect", duration: 999999 })

	}



	const ConnectToVPN = async (a, ar) => {

		try {

			if (!STORE.ActiveRouterSet(props.state)) {
				return
			}

			props.toggleLoading({ logTag: "connect", tag: "CONNECT", show: true, msg: "Connecting you to VPN " + a.Tag, includeLogs: true })

			if (!user.DeviceToken) {
				LogOut()
				return
			}

			let method = undefined
			if (props.state?.ActiveAccessPoint) {
				method = Switch
			} else {
				method = Connect
			}

			// console.log("CONNECTING TO:", a.Tag, a.GROUP, a.ROUTERID)
			// console.log("SECOND ROUTER:", ar.GROUP, ar.ROUTERID)

			let ConnectForm = {
				UserID: user._id,
				DeviceToken: user.DeviceToken.DT,

				GROUP: ar.GROUP,
				ROUTERID: ar.ROUTERID,

				XGROUP: a.Router.GROUP,
				XROUTERID: a.Router.ROUTERID,
				DEVICEID: a.DEVICEID,
			}

			if (a.Networks) {
				ConnectForm.Networks = a.Networks
			}

			method(ConnectForm).then((x) => {
				if (x.Code === 401) {
					LogOut()
				}

				if (x.Err) {
					props.toggleError(x.Err)
				} else {
					if (x.Code === 200) {

						STORE.Cache.Set("connected_quick", "XX")

						props.showSuccessToast("Connected to VPN " + a.Tag, undefined)

					} else {
						props.toggleError(x.Data)
					}
				}

			}).catch((e) => {
				console.dir(e)
				props.toggleError("Unknown error, please try again in a moment")
			})

		} catch (error) {
			console.dir(error)
		}

		props.toggleLoading(undefined)

	}

	const NavigateToEditAP = (id) => {
		navigate("/accesspoint/" + id)
	}
	const NavigateToCreateAP = () => {
		navigate("create/accesspoint")
	}

	let user = STORE.GetUser()
	if (!user) {
		return (<Navigate to={"/login"} />)
	}



	const RenderServer = (ap, ar, editButton, isConnected) => {
		let method = undefined
		if (isConnected) {
			method = undefined
		} else {
			method = ConfirmConnect
		}

		let connected = false
		if (props.state?.ActiveAccessPoint?._id == ap._id) {
			connected = true
		}

		if (!ap.Online) {
			return (
				<>
					<div className={`server`} onClick={() => NavigateToEditAP(ap._id)}>

						<div className="item tag tag-offline" >{ap.Tag} </div>
						<div className="item offline-text" >
							{`( OFFLINE )`}

						</div>
						<div className="item x3"></div>
						{/* <div className="item x3"></div> */}
						<div className="item x3"></div>
						<div className="item x3"> </div>
						<div className="item x3"></div>
						<div className="item x3"></div>
					</div>
				</>
			)

		}

		let country = "icon"
		if (ap.GEO !== undefined && ap.GEO.Country !== "") {
			country = ap.GEO.Country.toLowerCase()
		} else if (ap.Country !== "") {
			country = ap.Country.toLowerCase()
		}


		return (
			<>
				<div className={`server ${isConnected ? `is-connected` : ``}`} onClick={() => method(ap, ar)} >

					{connected &&
						<div className="item tag"  >
							<EnterIcon className="icon"></EnterIcon>
							{ap.Tag}
						</div>
					}
					{!connected &&
						<div className="item tag"  >
							{ap.Tag}</div>
					}

					<div className="item country" >
						{country !== "icon" &&
							<>
								<img
									className="country-flag"
									src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + country + ".svg"}
								// src={"/src/assets/images/flag/" + ap.GEO.Country.toLowerCase() + ".svg"}
								/>
								<div className="text">
									{country.toUpperCase()}
								</div>
							</>
						}
						{country === "icon" &&

							<>
								<DesktopIcon className="country-temp" height={23} width={23}></DesktopIcon>
								<div className="text">
									Private
								</div>
							</>
						}

					</div>

					{ap.Router &&
						<>
							<div className="item x3">{ap.Router.Score}</div>
							<div className="item x3">{ap.Router.AvailableSlots} / {ap.Router.Slots}</div>
							<div className="item x3">{ap.Router.AvailableMbps / 1000}</div>
							<div className="item x3">{ap.Router.AIBP} / {ap.Router.AEBP}</div>
							<div className="item x3">{ap.Router.CPUP}</div>
							<div className="item x3">{ap.Router.RAMUsage}</div>
						</>
					}
				</div>
			</>
		)
	}

	let AccessPoints = []
	let PrivateAccessPoints = []

	if (props?.state?.PrivateAccessPoints) {

		if (filter && filter !== "") {


			props.state.PrivateAccessPoints.map(r => {

				let filterMatch = false
				if (r.Tag?.toLowerCase().includes(filter)) {
					filterMatch = true
				}

				if (filterMatch) {
					PrivateAccessPoints.push(r)
				}

			})

		} else {
			PrivateAccessPoints = props.state.PrivateAccessPoints
		}

	}

	if (props?.state?.AccessPoints) {

		if (filter && filter !== "") {

			props.state.AccessPoints.map(r => {

				let filterMatch = false
				if (r.Tag?.toLowerCase().includes(filter)) {
					filterMatch = true
				} else if (r.GEO?.Country?.toLowerCase().includes(filter)) {
					filterMatch = true
				} else if (r.GEO?.CountryFull?.toLowerCase().includes(filter)) {
					filterMatch = true
				}

				if (filterMatch) {
					AccessPoints.push(r)
				}

			})

		} else {
			AccessPoints = props.state.AccessPoints
		}

	}

	const RenderSimpleServer = (ap, ar) => {
		let country = "icon"
		if (ap.GEO !== undefined && ap.GEO.Country !== "") {
			country = ap.GEO.Country.toLowerCase()
		} else if (ap.Country !== "") {
			country = ap.Country.toLowerCase()
		}

		let connected = false
		if (props.state?.ActiveAccessPoint?._id == ap._id) {
			connected = true
		}

		let method = function(x, y) {
			props.toggleError("You are already connected to this VPN")
		}
		if (!connected) {
			method = ConfirmConnect
		}


		return (
			<div className={`item ${connected ? "connected" : ""}`} onClick={() => method(ap, activeR)}>

				{country !== "icon" &&
					<>
						<img
							className="flag"
							src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + country + ".svg"}
						/>

					</>
				}
				{country === "icon" &&
					<div className="icon">
						<DesktopIcon className="icon" height={"auto"} width={"auto"}></DesktopIcon>

					</div>
				}


				<div className="info">
					<div className="tag">
						{ap.Tag}
					</div>
					<div className="score">Quality Score: {ap.Router.Score}</div>

				</div>





			</div>)
	}

	let activeR = props.state?.ActiveRouter

	if (!props.advancedMode) {
		return (
			<div className="server-wrapper">

				<div className="search-wrapper">
					<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
					<input type="text" className="search" onChange={updateFilter} placeholder="Search .."></input>
				</div>


				{(AccessPoints.length < 1 && PrivateAccessPoints.length < 1 && filter == "") &&
					<Loader
						className="spinner"
						loading={true}
						color={"#20C997"}
						height={100}
						width={50}
					/>
				}

				<div className="simple-list">

					{PrivateAccessPoints.map((ap) => {
						return RenderSimpleServer(ap, activeR)
					})}

				</div>

				<div className="simple-list">

					{AccessPoints.map(ap => {
						return RenderSimpleServer(ap, activeR)
					})}
				</div>

			</div >
		)
	}

	return (
		<div className="server-wrapper" >

			<div className="search-wrapper">
				<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
				<input type="text" className="search" onChange={updateFilter} placeholder="Search .."></input>
			</div>

			{activeR &&
				<div className="advanced-list advanced-list-bottom-margin" >

					<div className="header">
						<div className="title tag">Tag</div>
						<div className="title country">Location</div>
						<div className="title x3">QoS
							<span className="tooltiptext">{STORE.VPN_Tooltips[0]}</span>
						</div>
						<div className="title x3">Slots
							<span className="tooltiptext">{STORE.VPN_Tooltips[1]}</span>
						</div>
						<div className="title x3">Gbps
							<span className="tooltiptext">{STORE.VPN_Tooltips[5]}</span>
						</div>
						<div className="title x3">AB %
							<span className="tooltiptext">{STORE.VPN_Tooltips[2]}</span>
						</div>
						<div className="title x3">CPU %
							<span className="tooltiptext">{STORE.VPN_Tooltips[3]}</span>
						</div>
						<div className="title x3">RAM %
							<span className="tooltiptext">{STORE.VPN_Tooltips[4]}</span>
						</div>
					</div>


					{(AccessPoints.length < 1 && PrivateAccessPoints.length < 1 && filter == "") &&
						<Loader
							className="spinner"
							loading={true}
							color={"#20C997"}
							height={100}
							width={50}
						/>
					}

					{PrivateAccessPoints.length > 0 &&
						PrivateAccessPoints.map(ap => {
							return RenderServer(ap, activeR, false, false)
						})
					}

					{AccessPoints.length > 0 &&
						AccessPoints.map(ap => {
							return RenderServer(ap, activeR, false, false)
						})
					}

				</div>
			}

		</div >
	);
}

export default Dashboard;

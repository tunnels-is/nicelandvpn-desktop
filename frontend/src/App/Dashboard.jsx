import { useNavigate, Navigate } from "react-router-dom";
import React, { useEffect, useState } from "react";

import {
	ChevronDownIcon,
	ChevronUpIcon,
	DesktopIcon,
	MagnifyingGlassIcon,
} from "@radix-ui/react-icons";

import Loader from "react-spinners/ScaleLoader";
import toast from 'react-hot-toast';
import STORE from "../store";
import API from "../api";
import CustomSelect from "./Select";
import dayjs from "dayjs";

const Dashboard = (props) => {

	const navigate = useNavigate();

	const [filter, setFilter] = useState("");
	const [queryFilter, setQueryFilter] = useState();
	const [comparison, setComparison] = useState()
	const [nodes, setNodes] = useState([])
	const [timeout, setTimeout] = useState(0)
	const [openNodes, setOpenNodes] = useState(new Map())

	const toggleNode = (id) => {
		let e = openNodes
		let isOpen = e.get(id)
		if (!isOpen) {
			e.set(id, true)
		} else {
			e.set(id, false)
		}
		setOpenNodes(new Map(e))
	}
	//
	// const openNode = (id) => {
	// 	let e = openNodes
	// 	e.set(id, true)
	// 	setOpenNodes(new Map(e))
	// }
	//
	// const closeNode = (id) => {
	// 	let e = openNodes
	// 	e.set(id, false)
	// 	setOpenNodes(new Map(e))
	// }

	const inputKeyDown = (k) => {
		if (k.keyCode === 13) {
			apiSearch()
		}
	}

	useEffect(() => {
		const newdate = dayjs().subtract(1, 'day').unix()
		setTimeout(newdate)
	}, [])

	const apiSearch = async () => {

		let now = dayjs().unix()
		let diff = now - timeout

		console.log("SEARTCH QUERY")
		console.log(timeout)
		console.log(now)
		console.log("DIFF", diff)
		console.dir(queryFilter)
		console.dir(comparison)

		if (filter === "") {
			props.toggleError("please enter a search term")
			return
		}
		if (now - timeout < 4) {
			props.toggleError("you can search again in " + diff + " seconds")
			return
		}

		let FR = {
			Path: "v3/node/search",
			Method: "POST",
			Timeout: 10000,
			JSONData: [{
				Comparison: comparison.key,
				Key: queryFilter.key,
				Value: filter,
			}]
		}

		setTimeout(dayjs().unix())

		props.toggleLoading({ tag: "SEARCH", show: true, msg: "searching ..." })

		try {
			let resp = await API.method("forwardToController", FR)
			if (resp === undefined) {
				throw "no response from api"
			}

			if (resp.status === 200) {
				setNodes(resp.data)
			} else if (resp.status === 204) {
				props.toggleError("nothing found during search")
			} else {
				props.toggleError(resp.data)
			}

		} catch (error) {
			console.dir(error)
			props.toggleError("unknown error, please try again in a moment")
		}

		props.toggleLoading(undefined)

	}

	const updateFilter = (event) => {
		setFilter(event.target.value)
	}

	let user = STORE.GetUser()
	if (!user) {
		return (<Navigate to={"/login"} />)
	}

	const LogOut = () => {
		props.toggleError("You have been logged out")
		STORE.Cache.Clear()
	}


	const ConfirmConnect = (c) => {

		toast.success((t) => (
			<div className="notification-frame">

				<div className="title">
					Your are connecting to
				</div >

				<div className="subtitle">
					{c}
					<img className="flag" src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + c.toLowerCase() + ".svg"} />
				</div>


				<div className="button-wrapper">
					<button className="cancel" onClick={() => {
						toast.dismiss(t.id)
						ConnectToVPN(c)
					}}>
						Connect
					</button>

					<button className="exit" onClick={() =>
						toast.dismiss(t.id)}>
						Cancel
					</button>
				</div>

			</div >
		), { id: "connect", duration: 999999 })

	}



	const ConnectToVPN = async (c) => {

		try {

			props.toggleLoading({ logTag: "connect", tag: "CONNECT", show: true, msg: "Connecting...", includeLogs: true })

			let user = STORE.GetUser()
			if (!user.DeviceToken) {
				LogOut()
				return
			}

			let method = "connect"
			let connectionRequest = {}
			connectionRequest.UserID = user._id
			connectionRequest.DeviceToken = user.DeviceToken.DT
			connectionRequest.Country = c


			let resp = await API.method(method, connectionRequest)
			if (resp === undefined) {
				props.toggleError("Unknown error, please try again in a moment")
			} else {
				if (resp.status === 401) {
					LogOut()
				} else if (resp.status === 200) {
					props.showSuccessToast("connection ready!", undefined)
				} else if (resp.data) {
					console.log("HAHSJDHAKSJHDKHASDKJ")
					console.dir(resp.data)
					props.toggleError(resp.data)
				} else {
					props.toggleError("Unknown error, please try again in a moment")
				}
			}


		} catch (error) {
			console.dir(error)
		}

		props.toggleLoading(undefined)

	}

	const NavigateToEditAP = (id) => {
		navigate("/accesspoint/" + id)
	}



	const RenderNode = (node) => {


		if (!node.Online) {
			return (
				<div className={`node`}
				// onClick={() => NavigateToEditAP(node._id)}
				>
					<div className="item tag tag-offline" >{node.Tag} </div>
				</div >
			)
		}

		let country = "icon"
		if (node.Country !== "") {
			country = node.Country.toLowerCase()
		}
		let lastOnline = dayjs(node.LastOnline)
		let now = dayjs().unix()
		let lastOnlineUnix = lastOnline.unix()
		let warningClass = "green"
		if (!node.TIME_PARSED) {
			if (now - lastOnlineUnix > 60) {
				warningClass = "orange"
			} else if (now - lastOnlineUnix > 120) {
				warningClass = "red"
			}
		}
		node.TIME_PARSED = true


		return (
			<>
				<div className={`node`} >
					<div className="item icon" onClick={() => toggleNode(node._id)} >
						{!openNodes.get(node._id) &&
							<ChevronDownIcon
								className="green"
								height={30}
								width={30}>
							</ChevronDownIcon>
						}
						{openNodes.get(node._id) &&
							<ChevronUpIcon
								className="orange"
								height={30}
								width={30}>
							</ChevronUpIcon>
						}
					</div>

					<div className="item tag">{node.Tag}</div>
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
								<div className="text green">
									Private
								</div>
							</>
						}

					</div>

					<div className="item slots">
						Slots
						<span className="green">{node.Slots}</span>
					</div>

					<div className="item slots">
						Sessions
						<span className="green">{node.Sessions} </span>
					</div>

					<div className={`item time`}>
						Ping
						<span className={warningClass}>{lastOnline.format('HH:mm:ss')} </span>
					</div>

				</div>

				<div className={`node-details ${openNodes.get(node._id) ? "open-details" : "closed-details"}`}>
					{Object.keys(node).map((k) => {
						if (node[k] && typeof node[k] !== "object" && typeof node[k] !== "array") {
							return (
								<div className="item">
									<div className="key">{k}</div>
									<div className="value">{String(STORE.formatNodeKey(k, node[k]))}</div>
								</div>
							)
						}
					})}
				</div>
			</>
		)
	}

	// let Nodes = []
	// let PrivateNodes = []
	//
	// if (props?.state?.PrivateNodes) {
	//
	// 	if (filter && filter !== "") {
	//
	//
	// 		props.state.PrivateNodes.map(r => {
	//
	// 			let filterMatch = false
	// 			if (r.Tag?.toLowerCase().includes(filter)) {
	// 				filterMatch = true
	// 			}
	//
	// 			if (filterMatch) {
	// 				PrivateNodes.push(r)
	// 			}
	//
	// 		})
	//
	// 	} else {
	// 		PrivateNodes = props.state.PrivateNodes
	// 	}
	//
	// }
	//
	// if (props?.state?.Nodes) {
	//
	// 	if (filter && filter !== "") {
	//
	// 		props.state.Nodes.map(r => {
	//
	// 			let filterMatch = false
	// 			if (r.Tag?.toLowerCase().includes(filter)) {
	// 				filterMatch = true
	// 			} else if (r.Country?.toLowerCase().includes(filter)) {
	// 				filterMatch = true
	// 			} else if (r.CountryFull?.toLowerCase().includes(filter)) {
	// 				filterMatch = true
	// 			}
	//
	// 			if (filterMatch) {
	// 				Nodes.push(r)
	// 			}
	//
	// 		})
	//
	// 	} else {
	// 		Nodes = props.state.Nodes
	// 	}
	//
	// }

	const RenderCountry = (c) => {
		console.log("COUNTRY")
		let country = c.toLowerCase()
		return (
			<div className={`country`} onClick={() => ConfirmConnect(c)}>
				<img className="flag" src={"https://raw.githubusercontent.com/tunnels-is/media/master/nl-website/v2/flags/" + country + ".svg"}
				/>
			</div>
		)
	}


	return (
		<div className="server-wrapper" >

			{!props.advancedMode &&
				<div className="countries">
					{!props.state.AvailableCountries &&
						<Loader
							className="spinner"
							loading={true}
							color={"#20C997"}
							height={100}
							width={50}
						/>}


					{props.state?.AvailableCountries?.map((c) => {
						return RenderCountry(c)
					})}

				</div>
			}

			{props.advancedMode &&
				<>
					<div className="search">

						<div className="submit">
							<MagnifyingGlassIcon
								className="search-icon"
								onClick={() => apiSearch()}
								height={30}
								width={30}>
							</MagnifyingGlassIcon>

						</div>

						<CustomSelect
							className={"filters"}
							setValue={setQueryFilter}
							defaultOption={{ key: "Tag", value: "Tag" }}
							options={[
								{ value: "Tag", key: "Tag" },
								{ value: "IP", key: "IP" },
								{ value: "Country", key: "Country" },
								{ value: "Slots", key: "Slots" },
								{ value: "ID", key: "ID" },
							]}
						></CustomSelect>

						<CustomSelect
							className={"comparisons"}
							setValue={setComparison}
							defaultOption={{ key: "=", value: "=" }}
							options={[
								{ value: "=", key: "=" },
								{ value: ">", key: ">" },
								{ value: "<", key: "<" },
							]}
						></CustomSelect>

						<input
							something="tag, id, country, ip, mbps, slots"
							type="text"
							onKeyDown={(k) => inputKeyDown(k)}
							className="input"
							onChange={updateFilter}
							placeholder="search ..">
						</input>
					</div>


					<div className="nodes">
						{(nodes.length < 1 && filter == "") &&
							<Loader
								className="spinner"
								loading={true}
								color={"#20C997"}
								height={100}
								width={50}
							/>}


						{nodes?.map((n) => {
							return RenderNode(n)
						})}

					</div>
				</>
			}


		</div >
	);
}

export default Dashboard;

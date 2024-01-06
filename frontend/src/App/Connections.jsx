import { Navigate } from "react-router-dom";
import React, { useState } from "react";


import STORE from "../store";
import { DesktopIcon, MagnifyingGlassIcon, EnterIcon } from "@radix-ui/react-icons";
import API from "../api";


var xxxxx = [
	{
		Name: "Niceland",
		IPv4Address: "10.4.3.3",
		IPv6Address: "",
		IFName: "nvpnclient",
		MTU: 1500,
		TxQueueLen: 3000,
		Persistent: true,
		Routes: [
			{ Name: "default", Route: "0.0.0.0/0" },
		],
		RouterIndex: 5,
		ProxyIndex: 5,
		NodeID: "65917765c027bcc5643460cd",
		NodePrivate: "",
		AutoReconnect: false,
		DNS: ["1.1.1.1", "9.9.9.9"],
		RouterProtocol: "tcp",
		RouterPort: "443",
	},
	{
		Name: "Cloud Edge 1",
		IPv4Address: "10.4.3.3",
		IPv6Address: "",
		IFName: "ce1",
		MTU: 1500,
		TxQueueLen: 3000,
		Persistent: true,
		Routes: [
			{ Name: "app-network", Route: "172.17.1.0/24" },
			{ Name: "db-network", Route: "172.18.1.0/24" }
		],
		RouterIndex: 5,
		ProxyIndex: 5,
		NodeID: 0,
		NodePrivate: "",
		AutoReconnect: true,
		DNS: ["1.1.1.1", "9.9.9.9"],
		RouterProtocol: "tcp",
		RouterPort: "443",
	}
]


const Connections = (props) => {
	STORE.Cache.SetObject("connections", xxxxx)

	const [filter, setFilter] = useState("");
	const [connections, setConnections] = useState(STORE.Cache.GetObject("connections"))

	let cons = []

	if (connections) {

		if (filter && filter !== "") {
			connections.map(r => {

				let filterMatch = false
				Object.keys(r).forEach((key, item) => {
					if (r[key].toString().includes(filter)) {
						filterMatch = true
					}
				})

				if (filterMatch) {
					cons.push(r)
				}

			})

		} else {
			cons = connections
		}

	}


	const LogOut = () => {
		props.toggleError("You have been logged out")
		STORE.Cache.Clear()
	}

	const ConnectToVPN = async (connection) => {

		try {

			props.toggleLoading({ logTag: "connect", tag: "CONNECT", show: true, msg: "Initializing " + connection.Name, includeLogs: true })

			let user = STORE.GetUser()
			if (!user) {
				LogOut()
				return (<Navigate to={"/login"} />)
			}

			// let method = undefined
			// if (props.state?.ActiveAccessPoint) {
			// 	method = "switch"
			// } else {
			let method = "connect"
			// }


			let connectionRequest = { ...connection }
			connectionRequest.UserID = user._id
			connectionRequest.DeviceToken = user.DeviceToken.DT


			let x = await API.method(method, connectionRequest)
			if (x === undefined) {
				props.toggleError("Unknown error, please try again in a moment")
			} else {
				if (x.status === 401) {
					LogOut()
				}
				if (x.status === 200) {
					props.showSuccessToast(connection.Name + " initialized", undefined)
				} else {
					props.toggleError(x.data)
				}
			}

		} catch (error) {
			console.dir(error)
		}

		props.toggleLoading(undefined)


	}

	const updateDNS = (c, index, newValue) => {

		console.dir(index)
		console.dir(c)
		console.dir(newValue)

		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name !== c.Name) {
				return
			}

			con.DNS[index] = newValue
		})

		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])

	}


	const updateRoute = (c, type, key, newValue) => {

		// console.dir(id)
		console.dir(c)
		console.dir(newValue)

		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name !== c.Name) {
				return
			}

			con.Routes.forEach((route) => {

				if (type === "route") {
					if (route.Name === key) {
						route.Route = newValue
					}
				} else {
					if (route.Route === key) {
						route.Name = newValue
					}
				}

			})
		})

		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])

	}

	const toggle = (c, id) => {
		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name === c.Name) {
				con[id] = !con[id]
			}
		})
		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])
	}

	const inputChange = (c, id, newValue) => {
		let consx = [...connections]
		consx.forEach((con) => {
			if (con.Name === c.Name) {
				con[id] = newValue
			}
		})
		STORE.Cache.SetObject("connections", consx)
		setConnections([...consx])
	}

	const RenderConnection = (c) => {
		return (
			<div className="connection" key={c.Name} >
				<div className="name">
					{c.Name}
				</div>
				<div className="connect" onClick={() => ConnectToVPN(c)}>
					CONNECT
				</div>

				<div className="title">Tunnel</div>
				<div className="cell">IPv4
					<input
						className="value"
						id="IPv4Address"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.IPv4Address}
					/>
				</div>

				<div className="cell">MTU
					<input
						className="value"
						id="MTU"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.MTU}
					/>
				</div>

				<div className="cell">TXQueueLen
					<input
						className="value"
						id="TxQueueLen"
						onChange={(e) => inputChange(c, e.target.id, e.target.value)}
						value={c.TxQueueLen}
					/>
				</div>

				<div className="cell">AutoReconnect
					<div
						className={`value toggle special-fix ${c.AutoReconnect ? "on" : "off"}`}
						id="AutoReconnect"
						onClick={(e) => toggle(c, e.target.id)}
					>{c.AutoReconnect ? "enabled" : "disabled"}</div>
				</div>

				<div className="cell">Persistent
					<div
						className={`value toggle ${c.Persistent ? "on" : "off"}`}
						id="Persistent"
						onClick={(e) => toggle(c, e.target.id)}
					>
						{c.Persistent ? "yes" : "no"}
					</div>
				</div>

				<div className="title">Routes</div>
				{
					c.Routes.map((r) => {
						return (
							<div className="cell" key={r.Name}>
								<input
									className="value route-name"
									onChange={(e) => updateRoute(c, "name", r.Route, e.target.value)}
									value={r.Name}
								/>
								<input
									className="value route"
									onChange={(e) => updateRoute(c, "route", r.Name, e.target.value)}
									value={r.Route}
								/>
							</div>
						)
					})
				}

				<div className="title">DNS</div>
				<div className="cell">
					<input
						className="value dns0"
						onChange={(e) => updateDNS(c, 0, e.target.value)}
						value={c.DNS[0]}
					/>
					<input
						className="value dns1"
						onChange={(e) => updateDNS(c, 1, e.target.value)}
						value={c.DNS[1]}
					/>
				</div>

				<div className="title">( Entry ) Router</div>
				<div className="cell">Tag
					<select name="Router" id="Router" value={c.RouterIndex}>
						<option defaultChecked className="hidden" key={"none"} value={"none"}>Select Router</option>
						{props.state?.Routers?.map((r) => {
							if (r.ListIndex === c.RouterIndex) {
								return (
									<option className="hidden" key={r.Tag} value={r.ListIndex}>{r.Tag}</option>
								)
							} else {
								return (
									<option key={r.Tag} value={r.ListIndex}>{r.Tag}</option>
								)
							}
						})}
					</select>
				</div>

				{props.state?.Routers?.map((x) => {
					if (x.ListIndex === c.RouterIndex) {
						return (
							<>
								<div className="cell">IP
									<input className="value" disabled value={x.PublicIP} />
								</div>
								<div className="cell">Mbps
									<input className="value" disabled value={x.AvailableMbps} />
								</div>
								<div className="cell">MS
									<input className="value" disabled value={x.MS} />
								</div>
								<div className="cell">Availability(slots)
									<input className="value" disabled value={x.AvailableSlots} />
								</div>
							</>
						)
					}
				})}

				<div className="title">( Exit ) Node</div>
				<div className="cell">Tag

					<select name="Node" id="Node" value={c.NodePrivate === "" ? c.NodeIndex : c.NodePrivate} >
						<option defaultChecked className="hidden" key={"none"} value={"none"}>Select Node</option>
						{props.state?.PrivateNodes?.map((r) => {
							if (r._id === c.NodePrivate) {
								return (
									<option className="hidden" key={r.ListIndex} value={r.ListIndex}>{r.Tag} </option>
								)
							} else {
								return (
									<option key={r.ListIndex} value={r.ListIndex}>{r.Tag}</option>
								)
							}
						})}
						{props.state?.PrivateNodes?.length > 0 &&
							<option disabled key={"seperator"} value={"------"}>----------</option>
						}
						{props.state?.Nodes?.map((r) => {
							if (r.ListIndex === c.NodeIndex) {
								return (
									<option className="hidden" key={r.ListIndex} value={r.ListIndex}>{r.Tag}</option>
								)
							} else {
								return (
									<option key={r.ListIndex} value={r.ListIndex}>{r.Tag}</option>
								)
							}
						})}
					</select>
				</div>

				{props.state?.Nodes?.map((x) => {
					if (x.ListIndex === c.NodeIndex) {
						return (
							<>
								<div className="cell">IP
									<input className="value" disabled value={x.IP} />
								</div>
								<div className="cell">Mbps
									<input className="value" disabled value={x.AvailableMbps} />
								</div>
								<div className="cell">MS
									<input className="value" disabled value={x.MS} />
								</div>
								<div className="cell">Availability(slots)
									<input className="value" disabled value={x.AvailableSlots} />
								</div>
							</>
						)
					}
				})}
				{props.state?.PrivateNodes?.map((x) => {
					if (x._id === c.NodePrivate) {
						return (
							<>
								<div className="cell">IP
									<input className="value" disabled value={x.IP} />
								</div>
								<div className="cell">Mbps
									<input className="value" disabled value={x.AvailableMbps} />
								</div>
								<div className="cell">MS
									<input className="value" disabled value={x.MS} />
								</div>
								<div className="cell">Availability(slots)
									<input className="value" disabled value={x.AvailableSlots} />
								</div>
							</>
						)
					}
				})}

			</div >
		)
	}


	return (
		<div className="connection-wrapper"  >

			<div className="search-wrapper">
				<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
				<input type="text" className="search" onChange={(e) => setFilter(e.target.value)} placeholder="Search .."></input>
			</div>

			<div className="connections-flex">
				{cons.map((c) => {
					return RenderConnection(c)
				})}
			</div>
		</div >
	);
}

export default Connections;

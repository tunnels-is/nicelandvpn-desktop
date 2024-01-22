import React, { useEffect, useState } from "react";

// import { GetLogs } from "../../wailsjs/go/main/Service";
import { MagnifyingGlassIcon } from "@radix-ui/react-icons";
import API from "../api";

const Logs = (props) => {

	const [filter, setFilter] = useState("");
	const [data, setData] = useState({})
	const [timer, setTimer] = useState(1)
	const [timeoutMS, setTimeoutMS] = useState(2000)

	useEffect(() => {

		const to = setTimeout(async () => {

			let resp = await API.method("getLogs", {})
			if (resp === undefined) {

			} else {
				setData(resp.data)
			}
			// await GetLogs(logCount, filter).then((x) => {
			// 	if (x.Data) {
			// 		setData(x.Data)
			// 		setlogCount(x.Data.Content.length)
			// 	}
			//
			// }).catch((e) => {
			// 	console.dir(e)
			// })

			let t = timer + 1
			if (t > 10000) {
				t = 1
			}

			setTimer(t)

		}, timeoutMS)

		return () => { clearTimeout(to); }

	}, [timer])

	let logLines = []
	if (filter !== "") {

		data.forEach((line) => {
			if (line.includes(filter)) {
				logLines.push(line)
			}
		});

	} else {
		if (data.length > 0) {
			logLines = data
		}
	}

	return (
		<div className="logs-wrapper">


			<div className="search-wrapper">
				<MagnifyingGlassIcon height={40} width={40} className="icon"></MagnifyingGlassIcon>
				<input type="text" className="search" onChange={(e) => setFilter(e.target.value)} placeholder="Search .."></input>
			</div>

			<div className="logs-window custom-scrollbar">

				{logLines.map((line, index) => {
					return (
						<div className="line">
							<label className={""}>{line}</label>
						</div >
					)
				})}

			</div>
		</div >
	);


}

export default Logs;

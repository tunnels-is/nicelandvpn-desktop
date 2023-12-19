import axios from "axios"
var API = {
	async method(method, data) {
		try {
			let response = undefined
			let host = window.location.host
			host = host.split(":")
			if (data !== undefined) {
				response = await axios.post("http://" + host[0] + ":9999/v1/method/" + method, JSON.stringify(data), { headers: { "Content-Type": "application/json" } })
			} else {
				response = await axios.post("http://" + host[0] + ":9999/v1/method/" + method, {}, { headers: { "Content-Type": "application/json" } })
			}
			return response
		} catch (error) {
			console.dir(error)
			return undefined
		}
	},
}

export default API;

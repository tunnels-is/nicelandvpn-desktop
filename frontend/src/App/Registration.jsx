import React, { useState, useEffect } from "react";
import { useNavigate, Navigate } from "react-router-dom";

import { v4 as uuidv4 } from 'uuid';


import STORE from "../store";
import API from "../api";

const useForm = (props) => {
	const [inputs, setInputs] = useState({});
	const [errors, setErrors] = useState({});
	const [tokenLogin, setTokenLogin] = useState(false)
	const navigate = useNavigate();

	const GenerateToken = () => {
		if (!tokenLogin) {
			let token = uuidv4()
			setTokenLogin(true)
			errors["email"] = "SAVE THIS TOKEN!"
			setErrors({ ...errors })
			setInputs(inputs => ({ ...inputs, ["email"]: token }));
		} else {
			setTokenLogin(false)
			errors["email"] = ""
			setErrors({ ...errors })
			setInputs(inputs => ({ ...inputs, ["email"]: "" }));
		}
	}

	const HandleSubmit = async (event) => {

		let errors = {}
		let hasErrors = false

		if (!inputs["email"]) {
			errors["email"] = "Email/Username missing"
			hasErrors = true
		}

		if (inputs["email"]) {
			if (inputs["email"].length > 320) {
				errors["email"] = "Maximum 320 characters"
				hasErrors = true
			}

			if (!tokenLogin) {
				if (!inputs["email"].includes(".") || !inputs["email"].includes("@")) {
					errors["email"] = "Invalid email format"
					hasErrors = true

				}
			}
		}

		if (!inputs["password"]) {
			errors["password"] = "Password missing"
			hasErrors = true
		}
		if (!inputs["password2"]) {
			errors["password2"] = "Password confirm missing"
			hasErrors = true
		}

		if (inputs["password"] !== inputs["password2"]) {
			errors["password2"] = "Passwords do not match"
			hasErrors = true
		}

		if (inputs["password"]) {
			if (inputs["password"].length < 10) {
				errors["password"] = "Minimum 10 characters"
				hasErrors = true
			}
			if (inputs["password"].length > 255) {
				errors["password"] = "Maximum 255 characters"
				hasErrors = true
			}
		}

		if (hasErrors) {
			setErrors({ ...errors })
			return
		}

		props.toggleLoading({ tag: "REGISTER", show: true, msg: "Creating your account" })

		let x = await API.method("forwardToController", {
			Path: "v2/user/create",
			Method: "POST",
			JSONData: inputs,
			Timeout: 20000
		})

		if (x === undefined) {
			props.toggleError("Unknown error, please try again in a moment")
		} else {
			if (x.status === 200) {
				STORE.Cache.Set("default-email", inputs["email"])
				props.showSuccessToast("Your account has been created, please log in", undefined)
				setErrors({})
				navigate("/login");
			} else {
				props.toggleError(x.data)
				setErrors({})
			}

		}

		props.toggleLoading(undefined)
	}

	const HandleInputChange = (event) => {
		setInputs(inputs => ({ ...inputs, [event.target.name]: event.target.value }));
	}

	return {
		inputs,
		HandleInputChange,
		HandleSubmit,
		errors,
		navigate,
		tokenLogin,
		GenerateToken
	};
}

const Register = (props) => {

	const { inputs, HandleInputChange, HandleSubmit, errors, navigate, tokenLogin, GenerateToken } = useForm(props);

	const user = STORE.Cache.GetObject("user");
	if (user) {
		return (<Navigate replace to="/login" />)
	}

	const NavigateToLogin = () => {
		navigate("/login")
	}

	useEffect(() => {

		// window.addEventListener("keypress", function(event) {
		// 	if (event.key === "Enter") {
		// 		HandleSubmit()
		// 	}
		// });

	}, [])

	return (
		<div className="register-wrapper login-form-styles">

			{tokenLogin &&
				<>
					<div className="register-token-warning">
						WARNING
						<br />
						if you lose your account token we can not help you in any way to recover your account incase of a lost password. This token is the ONLY way to verify account ownership.
						<br />
					</div>
				</>
			}

			<div className="register-form" >

				{errors["email"] &&
					<div className="error email-error">
						{errors["email"]}<br />
					</div>
				}
				{errors["password"] &&
					<div className="error password-error">
						{errors["password"]}<br />
					</div>
				}
				{errors["password2"] &&
					<div className="error password2-error">
						{errors["password2"]}<br />
					</div>
				}

				{tokenLogin &&

					<input className="input token-input"
						autocomplete="off"
						placeholder={"Username / Token"}
						type="text"
						value={inputs["email"]}
						name="email"
						onChange={HandleInputChange} />
				}

				{!tokenLogin &&
					<input className="input email-input"
						autocomplete="off"
						placeholder={"Email"}
						type="email"
						value={inputs["email"]}
						name="email"
						onChange={HandleInputChange} />
				}




				<input className="input pass-input"
					placeholder={"Password"}
					type="password"
					value={inputs["password"]}
					name="password"
					onChange={HandleInputChange} />


				<input className="input pass2-input"
					type="password"
					placeholder={"Confirm Password"}
					value={inputs["password2"]}
					name="password2"
					onChange={HandleInputChange} />

				<input className="input code-input"
					type="text"
					placeholder={"code (optional)"}
					value={inputs["code"]}
					name="code"
					onChange={HandleInputChange} />


				<button className={`btn register-button`}
					onClick={HandleSubmit}
				>Register</button>

				{tokenLogin &&
					<button className={`btn token-button`}
						onClick={GenerateToken}
					>Use Email</button>
				}
				{!tokenLogin &&
					<button className={`btn token-button`}
						onClick={GenerateToken}
					>Use Token</button>
				}


			</div>

			<div className="login-options">
				<button className="button"
					onClick={() => NavigateToLogin()} >Back To Login</button>

			</div>

		</div >
	)
}

export default Register;

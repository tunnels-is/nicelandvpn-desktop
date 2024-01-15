
import { useEffect, useState } from "react";

const CustomSelect = (props) => {
	const [filterSelect, setFilterSelect] = useState({
		open: false,
		opt: { key: "", value: "" }
	})

	const filterState = (open, opt) => {
		setFilterSelect({ open: open, opt: opt })
		props.setValue(opt)
	}

	useEffect(() => {
		if (filterSelect.opt.value === "") {
			props.setValue(props.defaultOption)
			setFilterSelect({ open: false, opt: props.defaultOption })
		}
	}, [])

	return (
		<div className={`custom-select ${props.className}`} >

			<div className={`default `}
				onClick={() => filterState(!filterSelect.open, filterSelect.opt)}
				id={filterSelect.opt.value}>{filterSelect.opt.key}
			</div>

			<span className={`options ${filterSelect.open ? 'show' : 'hide'} `}>
				{props.options.map((opt) => {
					return (
						<div className={`opt`} id={opt.key}
							onClick={() => filterState(false, opt)}>
							{opt.value}
						</div>
					)
				})}
			</span>

		</div >
	)
}

export default CustomSelect;

import React, { useEffect, useState, useId } from "react";
import { useNavigate, Navigate } from "react-router-dom";
import STORE from "../store";
import ScreenLoader from "./ScreenLoader";
import Spinner from "./Spinner";

const Debug = (props) => {

  const [lines, setLines] = useState([])
  const [dfile, setDfile] = useState("")
  const [loading, setLoading] = useState(undefined)


  const RUN_DEBUG = async () => {


    setDfile("")
    setLines([])

    setLoading({ logTag: "", tag: "DEBUG", show: true, msg: "Debugging network settings...", includeLogs: false })

    try {
      // const response = await CLIENT.get("/debug");
      // let out = await response.data
      // console.dir(out)
      // setLines(out.Lines)
      // setDfile(out.File)
      // props.showSuccessToast("Debug complete", undefined)
    } catch (error) {
      console.dir(error)
      // props.toggleError(STORE.ParseResponseErrorMessage(error))
    }

    setLoading(undefined)
  }




  return (
    <div className="debug-wrapper">

      {loading &&
        <ScreenLoader loader={loading} toggleError={props.toggleError}></ScreenLoader>
      }

      <button className="debug-button"
        onClick={() => RUN_DEBUG()} >DEBUG</button>
      {dfile !== "" &&
        <input type="text" className="input search-bar" value={dfile} placeholder=""></input>
      }
      {dfile === "" &&
        <input type="text" className="input search-bar" value="" placeholder={"Click DEBUG to start debugging"}></input>

      }

      <div className="logs-window-bg"></div>
      <div className="logs-window custom-scrollbar">
        {lines.map(line => {
          return (<div className="line">{line}</div>)
        })}
      </div>
    </div >
  );

}

export default Debug;
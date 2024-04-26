import {useEffect, useRef, useState,useCallback} from 'react';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import Webcam from "react-webcam";
import { Button } from '@mui/material';
import io from 'socket.io-client'
import { abort } from 'process';
import Peer from 'simple-peer';




const socket = io("endpoint")

export default function SecondFactor({back}) {


    
    const connectionRef = useRef()
    const webcamRef = useRef(null);
    const mediaRecorderRef = useRef(null);
    const [capturing, setCapturing] = useState(false);
    const [recordedChunks, setRecordedChunks] = useState([]);
  
    

    const handleSendStream =() => {
        setCapturing(true);

        const peer = new Peer({
            initiator: true,
            trickle: false,
            stream: webcamRef.current.stream
        })

        peer.on("signal", (data) => {
			socket.emit("sendStreamToBackend", {
				signalData: data,
			})
		})

        socket.on("answer", (signal)=>{
            peer.signal(signal)
        })

        connectionRef.current = peer
        
    };


    var [colorNum, setColorNum] = useState(0);
    const [backgroundCol, setBackgroundColor] = useState("")
    const [stripCol, setstripCol] = useState("") 
    const [stripPos, setstripPos] = useState("")

    // const [colorList, setColorsList] = useState();
    // const [poslist, position] = useState()
   

    // useEffect(()=>{
    //     socket.on("disconnect", () => {
	// 		connectionRef.current.destroy()
	// 	})
    // })


    const colorList = ["#FF0000", "#0000FF", "#00FF00"];
    const poslist = ["10%", "20%", "30%", "40%"]
    useEffect(()=>{
        let colorNum = 0;
        let colorNum2=1;
        let posnum = 0;

        const changeColor = () => {
            colorNum=((colorNum+1)%3);
            colorNum2=((colorNum+1)%3);
            posnum = (posnum+1)%4;
            

            setBackgroundColor(colorList[colorNum]);
            setstripCol(colorList[colorNum2]);
            setstripPos(poslist[posnum]);
            
            console.log(colorList[colorNum]);
             // Call changeColor function again after 1 second
            console.log("changing a color");
            
        };
        const setinterval= setInterval(changeColor,500);
        return ()=>{clearInterval(setinterval)}
    
    },[])


    return (
     
        <>
            
            <Webcam style={{position:'absolute', left:0, right:0, top:0, bottom:0, margin:'auto', width:"100vw"}}/>

           

            <div style={{backgroundColor:backgroundCol, width:'100vw', height:'100vw', opacity:'70%'}}>
                <div style={{backgroundColor:stripCol, width:'100vw', height:'20%',position:'absolute',top:stripPos}}>
                
                </div>
            </div>
            <img src="oval.png" style={{position:'absolute', top:0, bottom:0, left:0, right:0, margin:'auto', height:'70%'}}/>
            {/* {capturing ? (
                <button onClick={handleStopCaptureClick}>Stop Capture</button>
            ) : (
                <button onClick={handleSendStream}>Start Capture</button>
            )} */}
            
            
        </>

      
    );
  };

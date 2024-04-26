import {useEffect, useRef, useState,useCallback} from 'react';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import Webcam from "react-webcam";
import { Button } from '@mui/material';
import io from 'socket.io-client'
import { abort } from 'process';
import Peer from 'simple-peer';




const websocket = new WebSocket("ws://localhost:8080/ws");


export default function SecondFactor({back}) {


    const websocket = new WebSocket("ws://localhost:8080/ws");

    websocket.onopen = e => {
      console.log("WebSocket connection established.");
      const message = {
        command: "ping",
        payload: { data: "Hello, this is a test message" },
      };
      websocket.send(JSON.stringify(message));
    };

    let remoteDescriptionSet = false;
    const candidateQueue = [];

    websocket.onmessage = async e => {
      const message = JSON.parse(e.data);
      console.log("Received message:", message);

      if (message.command === "IceAnswer") {
        const remoteDesc = new RTCSessionDescription(message.payload);
        console.log("Setting remote description...");
        await peerConnection.setRemoteDescription(remoteDesc);
        remoteDescriptionSet = true
        flushCandidateQueue()
      } else if (message.command === "IceCandidate") {
        const candidate = new RTCIceCandidate(message.payload);
        if (remoteDescriptionSet) {
          await peerConnection.addIceCandidate(candidate);
        } else {
          candidateQueue.push(candidate)
        }
      }
    };

    function flushCandidateQueue() {
      while (candidateQueue.length > 0) {
        const candidate = candidateQueue.pop()
        peerConnection.addIceCandidate(candidate)
      }
    }

    websocket.onerror = error => {
      console.error("WebSocket Error:", error);
    };

    websocket.onclose = event => {
      console.log("WebSocket Closed:", event.reason);
    };

    const configuration = { iceServers: [{ urls: "stun:stun.l.google.com:19302" }] };
    const peerConnection = new RTCPeerConnection(configuration);
    const dataChannel = peerConnection.createDataChannel("myDataChannel");

    peerConnection.onicecandidate = event => {
      if (event.candidate) {
        console.log("Sending new ICE candidate...");
        websocket.send(JSON.stringify({
          command: "IceCandidate",
          payload: event.candidate.toJSON()
        }));
      } else {
        console.log("ICE gathering complete.");
      }
    };

    async function startConnection() {
      console.log("Starting connection...");
      const offer = await peerConnection.createOffer();
      await peerConnection.setLocalDescription(offer);
      sendOffer();  
    }

    function sendOffer() {
      const message = {
        command: "IceOffer",
        payload: peerConnection.localDescription
      };
      websocket.send(JSON.stringify(message));
      console.log("Offer sent.");
    }

    peerConnection.onconnectionstatechange = event => {
      console.log("Connection state change:", peerConnection.connectionState);
    };

    dataChannel.onopen = () => {
      console.log("Data channel is open.");
    };

    dataChannel.onclose = () => {
      console.log("Data channel is closed.");
    };

    dataChannel.onerror = error => {
      console.error("Data channel error:", error);
    };

    async function sendMessage() {
      console.log("Sending message via data channel...");
      dataChannel.send("Hello World");
    }

    const startSend = ()=>{
        startConnection();
        
    }

    async function sendStream(){
        const localStream = await getUserMedia({video: true, audio: true});
        localstream.getTracks().forEach(track => {
            peerConnection.addTrack(track, localStream);
        });
    }
    
    const connectionRef = useRef()
    const webcamRef = useRef(null);
    const mediaRecorderRef = useRef(null);
    const [capturing, setCapturing] = useState(false);
    const [recordedChunks, setRecordedChunks] = useState([]);
  

    var [colorNum, setColorNum] = useState(0);
    const [backgroundCol, setBackgroundColor] = useState("")
    const [stripCol, setstripCol] = useState("") 
    const [stripPos, setstripPos] = useState("")


    const colorList = ["#FF0000", "#0000FF", "#00FF00"];
    const poslist = ["10%", "20%", "30%", "40%"];

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
            <button onClick={startSend}>Start Connection</button>
            
            
        </>

      
    );
  };
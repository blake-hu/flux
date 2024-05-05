import {useEffect, useRef, useState,useCallback} from 'react';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import Webcam from "react-webcam";
import { Button } from '@mui/material';
import io from 'socket.io-client'
import { abort } from 'process';
import Peer from 'simple-peer';
import { start } from 'repl';
import "./style.css"



export default function SecondFactor({back}) {
  
  const websocket = new WebSocket("ws://localhost:8080/ws");
  websocket.onopen = e => {
    console.log("WebSocket connection established.");
    const message = {
      command: "ping",
      payload: { data: "Hello, this is a test message" },
    };
    websocket.send(JSON.stringify(message));

    startConnection()
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
    } else if(message.command === "setBandColor"){
      
      setBandCol(message.payload.stripColor)
      setBandPos(message.payload.stripPosition)
      setBackgroundCol(message.payload.backgroundColor)
      confirmColorChange()
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
  initialize();
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

  // document.querySelector("#showVideo").addEventListener("click", e => initialize(e))

  async function initialize() {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: false,
      video: {width: 1280, height: 720}
    });
    attachVideoStream(stream)
  }

  function attachVideoStream(stream) {
    const videoElement = document.querySelector("video")
    window.stream = stream
    camVideo.current.srcObject = stream;
    peerConnection.addTrack(stream.getVideoTracks()[0], stream)
  }

  function closeInstruction(e){
    setInstructions(false)

    const message = {
      command: "readyForBandColor",
      payload: { data: "Ready" },
    };
    websocket.send(JSON.stringify(message));

    readyTimeout()
  }

  async function readyTimeout(){
    await sleep(10000)
    peerConnection.close()
    console.log("connection Closed")

  }

  async function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

  


  const [backgroundCol, setBackgroundCol] = useState()
  const [bandCol,setBandCol] = useState()
  const [bandPos, setBandPos] = useState()
  const [instructions, setInstructions] = useState(true)
  const [displayData, setDisplayData] = useState(null)
  const camVideo = useRef()
  const [nextData, setNextData] = useState()


  // useEffect(()=>{
  //   if(displayData === null){
  //     return
  //   }
  //   let size = displayData.length;
  //   let current = 0
  //   const interval = setInterval(() => {
  //     if (current=== size) { 
  //       return () => clearInterval(interval);
  //     }

  //     const newColor = colors[Math.floor(Math.random() * colors.length)]; // Choose a random color
      
  //   }, 20); // Change color every 5 seconds

  //   return () => clearInterval(interval);

  // }, [displayData])

  async function confirmColorChange() {
    console.log("Sending message via data channel...");
    const message = {
      timestamp: Date.now()
    }
    dataChannel.send(JSON.stringify(message));
  }
  
    return (
        <>            

            <video id="gum-local" playsInline autoPlay ref={camVideo} style={{position:'absolute', left:0, right:0, top:0, bottom:0, margin:'auto', width:"100vw"}} />
            <div style={{backgroundColor:backgroundCol, width:'100vw', height:'100vw', opacity:'70%'}}>
                <div style={{backgroundColor:bandCol, width:'100vw', height:'20%',position:'absolute',top:bandPos}}>
                
                </div>
            </div>
            <img src="oval.png" style={{position:'absolute', top:0, bottom:0, left:0, right:0, margin:'auto', height:'70%'}}/>

            {instructions &&
                <div className='instructionPopup'>
                  <h1>
                    Instructions for Facial Authentication
                  </h1>
                  <ul>
                    <li>
                      Ensure your face is brightly lit
                    </li>
                    <li>
                      Position your face according to the outline
                    </li>
                    <li>
                      Hold Still until flashing is complete
                    </li>
                  </ul>
                <button onClick={closeInstruction}>I'm Ready</button>
              </div>
          }
            
        </>
    );
  };

  
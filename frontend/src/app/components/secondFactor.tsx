import * as React from 'react';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import Webcam from "react-webcam";
import { Button } from '@mui/material';

export default function SecondFactor({back}) {
    return (
        <div>
            <div className="webCamDiv">
                <Webcam className="webCam"/>
            </div>
            <Button className='backButton' onClick ={()=>back()}>back</Button>
        </div>
    );
  }
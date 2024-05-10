import * as React from 'react';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import { Button } from '@mui/material';

export default function Failure() {
    return (
        <div className="failureOutlineBorder">
            <h1><b>Authentication</b></h1>
            <h1><b>Failed!</b></h1>
            <Button className="auth" style={{marginTop:"70px", backgroundColor:"red"}} onClick ={()=>window.location.reload()}>Retry</Button>
           
        </div>
    );
  }
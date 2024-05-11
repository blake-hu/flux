import * as React from 'react';
import Box from '@mui/material/Box';
import TextField from '@mui/material/TextField';
import Button from '@mui/material/Button';
import "./style.css";

export default function FirstFactor({next, email, setEmail}) {

    const handleSubmit = (event: { preventDefault: () => void; }) => {
      event.preventDefault();
    };

    const handleEmailChange = (event) => {
      setEmail(event.target.value);
    };
  
  
    return (
      
      <div className="outlineBorder">
        <Box
            component="form"
            autoComplete="off"
            onSubmit={handleSubmit}
          >
          <h1 className='signIn'>Sign In</h1>
          <h2 className='with'>with <b>Flux</b></h2>
          <div className="TextField">
            <TextField className='textfield' required id="uid" label="Email" variant="outlined" value={email} onChange={handleEmailChange}/>
          </div>
          <div className="TextField">
            <TextField className='textfield' required id="password" label="Password" variant="outlined" />
          </div>
          <a href="/recovery" className="forgotPassword">Forgot your password?</a>
          <div className='registerauthenticate'>
            <a href="/register" className="register">Register</a>
            <Button className="auth" onClick ={()=>next()}>Authenticate</Button>
          </div>
        </Box>
      </div>
    );
  }
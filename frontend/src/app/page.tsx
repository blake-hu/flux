"use client";
import FirstFactor from "./components/firstFactor";
import SecondFactor from "./components/secondFactor";
import Success from "./components/success";
import {ReactElement, useEffect, useRef, useState} from "react"

export default function Home() {

  const [currentStepNumber, setCurrentStepNumber] = useState(0)

  function next(){
    setCurrentStepNumber(i=>{
        if(i>=3){
            return i
        }
        return i+1
    })
  }

  function back(){
    setCurrentStepNumber(i=>{
        if(i<=0){
            return i
        }
        return i-1
    })
  }

  function goTo(number:number){
      setCurrentStepNumber(number)
  }

  const EMAIL_REGEX= /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,4}$/i;

  // one uppercase, one lowercase, one digit, one special, 8-24 characters
  const PWD_REGEX = /^(>=.*[a-z])(?=.*[A-Z])(?=.*[0-9])(?=.*[!@#$%]).{8,24}$/;
 
  const [email, setEmail] = useState('');
  const [pwd, setPwd] = useState('');

  const [validEmail, setValidEmail] = useState(false);
  const [validPwd, setValidPwd] = useState(false);

  const [firstSuccess, setFirstSuccess] = useState(false);

  

  useEffect(()=>{
    const result = EMAIL_REGEX.test(email);
    setValidEmail(result);
  },[email]);

  useEffect(()=>{
    const result = PWD_REGEX.test(pwd);
    setValidEmail(result);
  },[email]);

  const elements = [<FirstFactor next = {next}/>, <SecondFactor back={back}/>, <Success back={back}/>]
  return (
    <div>
      {elements[currentStepNumber]}
    </div>
  );
}

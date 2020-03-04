import React, { Component } from 'react';
import './App.css';

class App extends Component {
  constructor(props){
    super(props)
    const WS_BACKEND_URL = process.env.REACT_APP_WS_BACKEND_URL
    this.socket = new WebSocket(`${WS_BACKEND_URL}/ws/darkKitchenState`);    

    this.state = {
      shelves: {
        "hot": [],
        "cold": [],
        "frozen": [],
        "overflow": [],
      },
      wastedOrdersDecay: 0,
      wastedOrdersNoSpace: 0,
      output: "Not Connected",
      minDriverDelay: "2",
      maxDriverDelay: "8",
      timeUnits: "1000",
      poissionRateParam: "3.25",
    }    
  

    this.renderShelf.bind(this)
    this.renderOrder.bind(this)

    this.openSockets()    
  }

  openSockets() {
    this.socket.onopen = () => {
      this.setState({ output: "Connected\n" })
    };
  
    this.socket.onmessage = (e) => {
      let jsonData = JSON.parse(e.data)
      let wastedOrdersDecay = jsonData["wastedOrdersDecay"]
      delete jsonData["wastedOrdersDecay"]

      let wastedOrdersNoSpace = jsonData["wastedOrdersNoSpace"]
      delete jsonData["wastedOrdersNoSpace"]      

      this.setState({ shelves: jsonData, wastedOrdersDecay: wastedOrdersDecay, wastedOrdersNoSpace: wastedOrdersNoSpace  })
    };        
  }

  renderOrder(order, idx) {
    let tempColors = {
      "hot": "red",
      "cold": "lightblue",
      "frozen": "lightgray",
    }
    let tempLabel = (
      <div style={{ padding: 5, border: "1px solid black", background: tempColors[order.temp], color: "black", width: 40, display: "inline-block", marginBottom: 5, marginRight: 10 }}>
        {order.temp}
      </div>
    )
    return (
      <div>
        <b>{idx}</b> {tempLabel} {order.id} {order.name} - <b>{Math.floor(order.normalizedHealth * 100)}%</b>
      </div>
    )
  }

  renderShelf(shelfKey){
    return (
      <div style={{ minWidth: 600, maxWidth: 600, height: 300, fontSize: 12,  padding: 20, border: "1px solid gray", margin: 5, overflowY: "scroll" }}>
        <p style={{ borderBottom: "1px solid gray"}}><b>{shelfKey}</b></p>
        {this.state.shelves[shelfKey].map((order, idx) => {
          return (
            <div>
              {this.renderOrder(order, idx)}
            </div>
          )
        })}
      </div>
    )
  }

  render() {
    return (
      <div className="App">
        <h1> Dark Kitchen Shelf State</h1>
        {/* The status of the websocket connection to the server */}
        <h4> Status: {this.state.output} </h4>
        <div>
          <div style={{ display: "inline-block"}}>
              <p> Wasted Orders b/c of decay : {this.state.wastedOrdersDecay} </p>
              <p> Wasted Orders b/c no space left: {this.state.wastedOrdersNoSpace} </p>
              <h4> Shelves</h4>              
              <div style={{ background: "white", borderRadius: 4, color: "black" }}>
                {Object.keys(this.state.shelves).map((shelf) => {
                  return (
                    <div style={{ display: "inline-block"}}>
                      {this.renderShelf(shelf)}  
                    </div>
                  )
                })}
              </div>            
            </div>
        </div>   
      </div>
    );
  }
}

export default App;

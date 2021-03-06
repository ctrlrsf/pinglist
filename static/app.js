var formatLatency = function(microsecs) {
  return (microsecs / 1e6).toFixed(2) + "ms"
}

var formatStatus = function(statusInt) {
  switch (statusInt) {
    case 0:
      return "Unknown"
    case 1:
      return "Offline"
    case 2:
      return "Online"
  }
}



var HostList = React.createClass({
  render: function() {
    if (this.props.error === true) {
      return (
        <div>
        <HostListAutoUpdate />
        Error communicating with server.
        </div>
      );
    }


    var keyList = new Array()
    for (var key in this.props.data) {
      keyList.push(key)
    }

    if (keyList.length > 0) {
      var hostList = this.props.data

      var hostLiNodes = keyList.map(function (key) {
        var host = hostList[key]
        return (
          <Host key={host.address} address={host.address} description={host.description} status={host.status} latency={host.latency} />
        );
      });

      return (
        <div>
        <HostListAutoUpdate />
        {hostLiNodes}
        </div>
      );
      } else {
        return (
          <div>
            <HostListAutoUpdate />
            <div>Not monitoring any hosts. Add a new host above.</div>
          </div>
        );
      }
  }
});

var Host = React.createClass({
  handleDelete: function() {
    console.log("Deleting: " + this.props.address)

    var deleteUrl = apiUrl + "/" + this.props.address

    $.ajax({
      url: deleteUrl,
      type: "DELETE",
      success: function(result) {
        console.log("Deleted successfully.")
      }
    });
  },

  render: function() {
    var unknownColor = "#DDDD00"
    var offlineColor = "#FF0000"
    var onlineColor = "#00FF00"
    var statusColor

    switch (this.props.status) {
      case 0:
        statusColor = unknownColor
        break
      case 1:
        statusColor = offlineColor
        break
      case 2:
        statusColor = onlineColor
        break
    }

    var divStyle = {
      color: statusColor,
      fontWeight: "bold",
    }

    return (
      <div className="container" style={{width: "500px", marginLeft: "2em"}}>
        <div className="row">
          <div className="col-md-1">
            <a href="#" onClick={this.handleDelete}>&#x2716;</a>
          </div>
          <div className="col-md-4">
            {this.props.address}
          </div>
          <div className="col-md-3">
            {this.props.description}
          </div>
          <div className="col-md-2" style={divStyle}>
            {formatStatus(this.props.status)}
          </div>
          <div className="col-md-2">
            {formatLatency(this.props.latency)}
          </div>
        </div>
      </div>
    );
  }
});

var HostListAutoUpdate = React.createClass({
  getInitialState: function() {
    return {
      isChecked: false,
      savedInterval: 0
    };
  },

  componentDidMount: function() {
    this.handleChange()
  },

  handleChange: function(event) {
    var isChecked = !this.state.isChecked
    var savedInterval = this.state.savedInterval

    console.dir(this.state)

    if (isChecked === true) {
      console.log("Starting auto update..")
      savedInterval = setInterval(renderHostList, 5000)
    } else {
      console.log("Stopping auto update..")
      clearInterval(savedInterval)
    }

    // Update component state
    this.setState({
      isChecked: isChecked,
      savedInterval: savedInterval
    });
  },
  render: function() {
    return (
      <div>
        Auto-update: <input type="checkbox" checked={this.state.isChecked} onChange={this.handleChange} />
      </div>
    );
  }
});

var HostAddForm = React.createClass({
  handleSubmit: function(e) {
    e.preventDefault();
    var address = React.findDOMNode(this.refs.address).value.trim();
    var description = React.findDOMNode(this.refs.description).value.trim();

    React.findDOMNode(this.refs.address).value = ""
    React.findDOMNode(this.refs.description).value = ""


    if (! address) {
      return;
    }

    var putUrl = this.props.url + "/" + address

    var hostJson = {
      "address" : address,
      "description" : description,
    }

    $.ajax({
      url: putUrl,
      type: "PUT",
      contentType: "application/json",
      data: JSON.stringify(hostJson),
      dataType: "json",
      success: function(result) {
        console.log("Host added successfully.")
      }.bind(this)
    });

    return;
  },

  render: function() {
    return (
      <form className="hostAddForm" onSubmit={this.handleSubmit}>
        <input type="text" placeholder="IP or hostname" ref="address" />
        <input type="text" placeholder="Description" ref="description" />
        <input type="submit" value="Add host" />
      </form>
    );
  },
});

var apiUrl = "http://" + window.location.host + "/api/hosts"

function renderHostList() {
  $.ajax({
    url: apiUrl,
    success: function(result) {
      if (result !== null) {
        var hostList = result
        React.render(<HostList data={hostList} />, document.getElementById("hostlist"))
      }
    },
    error: function (xhr, status, error) {
      React.render(<HostList error={true} />, document.getElementById("hostlist"))
    },
    timeout: 5000
  });

}

React.render(<HostAddForm url={apiUrl} />, document.getElementById("hostform"))
renderHostList()

class App extends React.Component {
    render() {
        return (<Home />);
    }
}

class Home extends React.Component {
    constructor(props) {
        super(props);
        this.state = {}

        this.clean = this.clean.bind(this);
        this.flush = this.flush.bind(this);
    }

    flush() {
    }

    clean() {
        $.ajax(method: "DELETE")

    }

    render() {
        return (
            <div>
                <header className="masthead mb-4">
                    <div className="inner">
                        <h2 className="masthead-brand">YouTube TV Queue
                            <button onClick={this.clean} type="button" className="btn btn-sm ml-2 mr-1 btn-warning"><i className="fas fa-broom"></i></button>
                            <button onClick={this.flush} type="button" className="btn btn-sm ml-1 mr-1 btn-danger"><i className="fas fa-trash"></i></button>
                        </h2>
                    </div>
                </header>
                <div className="container cards">
                    <div className="row justify-content-md-center">
                        <NewVideoForm/>
                        <LoggedIn/>
                    </div>
                </div>
            </div>
        )
    }
}

class NewVideoForm extends React.Component {
    constructor(props) {
        super(props);
        this.state = {}

        this.add = this.add.bind(this);
        this.replace = this.replace.bind(this);
    }

    add() {
        let url = $("#new-video").val()
        $.ajax({
            method: "POST",
            url: "/api/queue",
            data: JSON.stringify({url: url}),
            dataType: "json",
            success: () => {
                $("#new-video").val("")
            }
        })
    }

    replace() {
        let url = $("#new-video").val()
        $.ajax({
            method: "PUT",
            url: "/api/queue",
            data: JSON.stringify({url: url}),
            dataType: "json",
            success: () => {
                $("#new-video").val("")
            }
        })
    }

    render() {
        return (
            <div className="form-group">
                <div className="input-group mb-3">
                    <input class="form-control form-control-lg rounded" type="text" placeholder="Add new video" id="new-video"/>
                    <button onClick={this.add} type="button" className="btn btn-success btn-lg"><i className="fas fa-plus"></i></button>
                    <button onClick={this.replace} type="button" className="btn btn-danger btn-lg"><i className="fas fa-play"></i></button>
                </div>
            </div>
        )
    }
}
    
    

class LoggedIn extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            vids: []
        }

        this.serverRequest = this.serverRequest.bind(this);
    }
    
    serverRequest() {
        $.getJSON("/api/queue", res => {
            this.setState({
                vids: res.reverse()
            });

            setTimeout(() => this.serverRequest(), 1000)
        });
    }

    componentDidMount() {
        this.serverRequest();
    }

    render() {
        return (
            <div className="row">
                {this.state.vids.map(function(vid, i){
                    return (<Vid key={i} vid={vid} />);
                })}

                {this.state.vids.length == 0 &&
                    <div className="card border-secondary mb-3 pb-0 col-md-12">
                        <div className="card-body">
                            <h4 className="card-title">No videos added to the queue yet</h4>
                        </div>
                    </div>
                }
            </div>
        )
    }
}

class Vid extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            liked: "",
            seeking: false
        }

        this.play = this.play.bind(this);
        this.pause = this.pause.bind(this);
        this.skip = this.skip.bind(this);
        this.delete = this.delete.bind(this);
        this.seek = this.seek.bind(this);
        this.pauseSeekUpdate = this.pauseSeekUpdate.bind(this);
    }
    
    play() {
        $.ajax({method:"PUT", url: "/api/player/play"})
    }

    pause() {
        $.ajax({method:"PUT", url: "/api/player/pause"})
    }

    skip() {
        $.ajax({method:"PUT", url: "/api/player/skip"})
    }

    delete() {
        $.ajax({method:"DELETE", url: "/api/queue", data: JSON.stringify({url: this.props.vid.url}), dataType: "json"})
    }
    
    pauseSeekUpdate() {
        this.state.seeking = true
    }

    seek(e) {
        let pct = $(e.target).val()
        $.ajax({method: "PUT", url: "/api/player/seek/"+pct.toString(), success: () => {
            this.state.seeking = false
        }})
    }

    render() {
        let classes = "card text-left mb-3"
        if (this.props.vid.playing) {
            classes = classes + " border-success playing bg-secondary"
        }

        if (!this.props.vid.playing && this.props.vid.played) {
            classes = classes + " played"
        }

        let progress = {width: this.props.vid.progress.toString()+"%"}

        if (!this.state.seeking) {
            $(".seekbar > input").val(this.props.vid.progress)
        }

        return (
            <div className={classes}>
            <div className="card-body">
            <h4 className="card-title">{this.props.vid.title}</h4>
            <h6 className="card-subtitle mb-2 text-muted"><a href={this.props.vid.url}>{this.props.vid.url}</a></h6>
            {/* <p class="card-text"><code><pre>{JSON.stringify(this.props.vid)}</pre></code></p> */}
            <div className="float-left mt-2 seekbar">
                <input type="range" class="custom-range" onMouseUp={this.seek} onMouseDown={this.pauseSeekUpdate}/>
            </div>
            <div className="float-right mt-2">
            {this.props.vid.playing == true && 
            <span>
            <button onClick={this.play} type="button" className="btn btn btn-primary"><i className="fas fa-play"></i></button>
            <button onClick={this.pause} type="button" className="btn btn btn-primary ml-1"><i className="fas fa-pause"></i></button>
            <button onClick={this.skip} type="button" className="btn btn btn-primary ml-1"><i className="fas fa-step-forward"></i></button>
            </span>
            }
            <button onClick={this.delete} type="button" className="btn btn btn-danger ml-1"><i className="fas fa-trash"></i></button>
            </div>
            </div>
            </div>
        )
    }
}

ReactDOM.render(<App />, document.getElementById('app'));
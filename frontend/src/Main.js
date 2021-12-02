import React, {Component} from 'react';
import {Button, Col, Container, Row} from "reactstrap";

import Peers from "./Peers";
import MessageInput from "./MessageInput";
import Messages from "./Messages";

import update from 'immutability-helper';


const UNAME = "UNAME";
const PEERS = "PEERS";
const MESSG = "MESSG";
const MESSG = "STORY"; // {cmd=strory messages[{from, datetime, content}]}

export default class Main extends Component {
    constructor(props, context) {
        super(props, context);

        this.state = {
            socket: null,
            iam: null,
            interlocutor: null,
            peers: {},
            messages: {},
        }
    }

    componentDidMount() {
        // еб-сокет для общения с бэком 
        let socket = new WebSocket("ws://" + document.location.hostname + (document.location.port ? ':' + document.location.port : '') + "/ws");

        this.setState({socket: socket}, () => {
            socket.onopen = function () {
                console.log("Соединение установлено.");
                socket.send(JSON.stringify({cmd: UNAME}));
                socket.send(JSON.stringify({cmd: PEERS}));
            };

            socket.onmessage = this.onMessage;
            socket.onclose = function (event) {
                if (event.wasClean) {
                    console.log('Соединение закрыто чисто');
                } else {
                    console.log('Обрыв соединения');
                }
                console.log('Код: ' + event.code + ' причина: ' + event.reason);
            };

            socket.onerror = function (error) {
                console.log("Ошибка " + error.message);
            };
        });
    }



    onMessage = (event) => {
        console.log("Получены данные " + event.data);
        let parsedMessage = JSON.parse(event.data);

        if (!parsedMessage.cmd) {
            console.error("something wrong with data");
            return;
        }

        this.handler(parsedMessage);
    };

    handler = (msgObj) => {
        switch (msgObj.cmd) {
            case UNAME : {
                console.log("Получены данные JOPA UNAME " + msgObj);
                this.setState({iam: {name: msgObj.name, id: msgObj.id}})
                break;
            }
            case PEERS : {
                console.log("Получены данные JOPA PEERS " + msgObj);

                let peers = {};
                msgObj.peers.forEach((p) => {
                    console.log("Получены данные JOPA PEERS p " + p);


                    let v = this.state.peers[p.id];
                    p.counter = v ? v.counter : 0;
                    peers[p.id] = p
                });
                this.setState({peers: peers});
                return;
            }
            case MESSG : {
                console.log("Получены данные JOPA MESSG" + msgObj);

                let peerId = "";
                let fromName = "";
                let counter = 0;

                if (msgObj.from === this.state.iam.id) {
                    // это наше сообщение
                    peerId = msgObj.to;
                    fromName = this.state.iam.name;
                } else {
                    // это сообщение от другого пира
                    peerId = msgObj.from;
                    let peer = this.state.peers[peerId];
                    if (peer) {
                        fromName = peer.name;
                        counter = peer.counter + 1;
                    } else {
                        fromName = msgObj.from.substr(0, 10);
                    }
                }

                let oldMessages = this.state.messages[peerId];
                if (!oldMessages) {
                    oldMessages = []
                }

                let message = {
                    date: new Date().toLocaleTimeString(['ru-RU', 'en-US'], {hour12: false}),
                    isMine: msgObj.from === this.state.iam.id,
                    from: fromName,
                    content: msgObj.content
                };

                oldMessages.push(message);
                console.log(oldMessages);
                this.setState({
                    peers: update(this.state.peers, {[peerId]: {counter: {$set: counter}}}),
                    messages: update(this.state.messages, {[peerId]: {$set: oldMessages}})
                });
                break;
            }
            default : {
                console.warn("Unknown cmd: " + msgObj.cmd)
            }
        }
    };


    updatePeers = () => {
        console.log("updatePeers JOPA REQUEST!!!!!!~!");

        this.state.socket.send(JSON.stringify({cmd: PEERS}));
    };

    sendMessage = (msg) => {
        let cmd = JSON.stringify({
            cmd: MESSG,
            from: this.state.iam.id,
            to: this.state.interlocutor.id,
            content: msg
        });
        this.state.socket.send(cmd);
    };

    selectPeer = (peer) => {
        this.setState({
            interlocutor: peer,
            peers: update(this.state.peers, {[peer.id]: {counter: {$set: 0}}}),
        })
    };


    render() {
        let interlocutorName = this.state.interlocutor ? " with " + this.state.interlocutor.name : "";
        return (
            <Container className={"d-flex h-100"} fluid>
                <Row className={"flex-fill flex-columns"}>
                    <Col xs={9} className={"messages d-flex align-content-end flex-column"}>
                        <Row className={"mb-auto pl-3 chat-header"}>
                            <h3>Chat {interlocutorName}</h3>
                        </Row>
                        <Row className={" scroll-on-overflow"}>
                            <Messages
                            messages={this.state.interlocutor ? (this.state.messages[this.state.interlocutor.id]): []}/>
                        </Row>
                        <Row className={"chat-footer"}>
                            <MessageInput interlocutor={this.state.interlocutor} onSendMessage={this.sendMessage}/>
                        </Row>
                    </Col>
                    <Col xs={3} className={"d-flex flex-column"}>
                        <Row className={"peer-header"}>
                            <Col>
                                <h4>Peers </h4>
                            </Col>
                        </Row>
                        <Row noGutters className={"scroll-on-overflow"}>
                            <Col>
                                <Peers peers={this.state.peers} onSelectPeer={this.selectPeer}/>
                            </Col>
                        </Row>
                        <Row className={"mt-auto peer-header"}>
                            <Col>
                                <Button color="light" size={"sm"} onClick={this.updatePeers}>update peers</Button>
                            </Col>
                        </Row>
                    </Col>
                </Row>
            </Container>
        )
    }
}
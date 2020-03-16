import * as React from 'react';
import {KeyboardEvent} from 'react';
import NodeStore from "app/stores/NodeStore";
import {inject, observer} from "mobx-react";
import FormControl from "react-bootstrap/FormControl";
import {FaucetStore} from "app/stores/FaucetStore";
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import Button from 'react-bootstrap/Button'
import InputGroup from "react-bootstrap/InputGroup";

interface Props {
    nodeStore?: NodeStore;
    faucetStore?: FaucetStore;
}

@inject("nodeStore")
@inject("faucetStore")
@observer
export class FaucetAddressInput extends React.Component<Props, any> {

    updateSend = (e) => {
        this.props.faucetStore.updateSend(e.target.value);
    };

    executeSend = (e: KeyboardEvent) => {
        if (e.key !== 'Enter') return;
        console.log(this.props.faucetStore.send_addr)
        this.props.faucetStore.sendReq();
    };

    btnExecuteSend = () => {
        console.log(this.props.faucetStore.send_addr)
        this.props.faucetStore.sendReq();
    };

    render() {
        let {send_addr, sending} = this.props.faucetStore;

        return (
            <React.Fragment>
                <Row className={"mb-3"}>
                    <Col>
                        <InputGroup className="mb-3">
                            <FormControl
                                placeholder="Address"
                                aria-label="Address"
                                aria-describedby="basic-addon1"
                                value={send_addr} onChange={this.updateSend}
                                onKeyUp={this.executeSend}
                                disabled={sending}
                            />
                        </InputGroup>
                    </Col>
                </Row>
                <Row className={"mb-3"}>
                    <Col>
                        <Button
                            variant="primary"
                            size="sm" block
                            onClick={this.btnExecuteSend}
                            value={send_addr}
                            disabled={sending}>
                           Send 
                        </Button>
                    </Col>
                </Row>
            </React.Fragment>
        );
    }
}

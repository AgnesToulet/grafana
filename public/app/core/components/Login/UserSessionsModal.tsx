import React, { PureComponent } from 'react';
import { UserSession } from 'app/types';
import { Button, Icon, Modal } from '@grafana/ui';
import { FormModel } from './LoginCtrl';

export interface Props {
  sessions: UserSession[];
  login: (data: FormModel) => void;
  formData: FormModel;
}

export class UserSessionsModal extends PureComponent<Props> {
  login(sessionId: number) {
    const { formData, login } = this.props;
    formData.tokenId = sessionId;
    login(formData);
  }

  render() {
    const { sessions } = this.props;
    return (
      <Modal title="Sessions" isOpen={true}>
        <h3 className="page-sub-heading">Select a session to log out in order to log in.</h3>
        <div className="gf-form-group">
          <table className="filter-table form-inline">
            <thead>
              <tr>
                <th>Last seen</th>
                <th>Logged on</th>
                <th>IP address</th>
                <th>Browser &amp; OS</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {sessions.map((session: UserSession, index) => (
                <tr key={index}>
                  <td>{session.seenAt}</td>
                  <td>{session.createdAt}</td>
                  <td>{session.clientIp}</td>
                  <td>
                    {session.browser} on {session.os} {session.osVersion}
                  </td>
                  <td>
                    <Button size="sm" variant="destructive" onClick={() => this.login(session.id)}>
                      <Icon name="power" />
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Modal>
    );
  }
}

export default UserSessionsModal;

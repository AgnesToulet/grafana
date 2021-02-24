import React, { PureComponent } from 'react';
import config from 'app/core/config';

import { updateLocation } from 'app/core/actions';
import { connect } from 'react-redux';
import { StoreState, UserSession } from 'app/types';
import { getBackendSrv } from '@grafana/runtime';
import { hot } from 'react-hot-loader';
import appEvents from 'app/core/app_events';
import { AppEvents } from '@grafana/data';

const isOauthEnabled = () => {
  return !!config.oauth && Object.keys(config.oauth).length > 0;
};

export interface FormModel {
  user: string;
  password: string;
  email: string;
  tokenId: number;
}
interface Props {
  routeParams?: any;
  updateLocation?: typeof updateLocation;
  children: (props: {
    isLoggingIn: boolean;
    changePassword: (pw: string) => void;
    isChangingPassword: boolean;
    skipPasswordChange: Function;
    login: (data: FormModel) => void;
    oauthLogin: (oauthName: string, sessionId?: number) => void;
    disableLoginForm: boolean;
    ldapEnabled: boolean;
    authProxyEnabled: boolean;
    disableUserSignUp: boolean;
    isOauthEnabled: boolean;
    loginHint: string;
    passwordHint: string;
    sessions: UserSession[];
    oauthName: string;
  }) => JSX.Element;
}

interface State {
  isLoggingIn: boolean;
  isChangingPassword: boolean;
  sessions: UserSession[];
  oauthName: string;
}

export class LoginCtrl extends PureComponent<Props, State> {
  result: any = {};
  constructor(props: Props) {
    super(props);
    const state = {
      isLoggingIn: false,
      isChangingPassword: false,
      sessions: [],
      oauthName: '',
    };

    if (config.loginError) {
      appEvents.emit(AppEvents.alertWarning, ['Login Failed', config.loginError]);

      if (config.concurrentSessions?.length > 0) {
        console.log('sessions: ', config.concurrentSessions);
        state.sessions = config.concurrentSessions;
      }

      if (config.oauthName) {
        state.oauthName = config.oauthName;
      }
    }

    this.state = state;
  }

  changePassword = (password: string) => {
    const pw = {
      newPassword: password,
      confirmNew: password,
      oldPassword: 'admin',
    };
    if (!this.props.routeParams.code) {
      getBackendSrv()
        .put('/api/user/password', pw)
        .then(() => {
          this.toGrafana();
        })
        .catch((err: any) => console.error(err));
    }

    const resetModel = {
      code: this.props.routeParams.code,
      newPassword: password,
      confirmPassword: password,
    };

    getBackendSrv()
      .post('/api/user/password/reset', resetModel)
      .then(() => {
        this.toGrafana();
      });
  };

  oauthLogin = (oauthName: string, sessionId?: number) => {
    this.setState({
      oauthName: oauthName,
    });

    let redirectUrl = config.appSubUrl + '/login/' + oauthName;
    if (sessionId) {
      redirectUrl += '?sessionId=' + sessionId;
    }

    window.location.href = redirectUrl;
  };

  login = (formModel: FormModel) => {
    this.setState({
      isLoggingIn: true,
    });

    getBackendSrv()
      .post('/login', formModel)
      .then((result: any) => {
        this.result = result;
        if (this.result.tokens?.length > 0) {
          this.showTokenModal(this.result.tokens);
        } else if (formModel.password !== 'admin' || config.ldapEnabled || config.authProxyEnabled) {
          this.toGrafana();
          return;
        } else {
          this.changeView();
        }
      })
      .catch(() => {
        this.setState({
          isLoggingIn: false,
        });
      });
  };

  changeView = () => {
    this.setState({
      isChangingPassword: true,
    });
  };

  showTokenModal = (sessions: UserSession[]) => {
    this.setState({
      sessions: sessions,
    });
  };

  toGrafana = () => {
    // Use window.location.href to force page reload
    if (this.result.redirectUrl) {
      if (config.appSubUrl !== '' && !this.result.redirectUrl.startsWith(config.appSubUrl)) {
        window.location.href = config.appSubUrl + this.result.redirectUrl;
      } else {
        window.location.href = this.result.redirectUrl;
      }
    } else {
      window.location.href = config.appSubUrl + '/';
    }
  };

  render() {
    const { children } = this.props;
    const { isLoggingIn, isChangingPassword, sessions, oauthName } = this.state;
    const { login, oauthLogin, toGrafana, changePassword } = this;
    const { loginHint, passwordHint, disableLoginForm, ldapEnabled, authProxyEnabled, disableUserSignUp } = config;

    return (
      <>
        {children({
          isOauthEnabled: isOauthEnabled(),
          loginHint,
          passwordHint,
          disableLoginForm,
          ldapEnabled,
          authProxyEnabled,
          disableUserSignUp,
          login,
          oauthLogin,
          isLoggingIn,
          changePassword,
          skipPasswordChange: toGrafana,
          isChangingPassword,
          sessions,
          oauthName,
        })}
      </>
    );
  }
}

export const mapStateToProps = (state: StoreState) => ({
  routeParams: state.location.routeParams,
});

const mapDispatchToProps = { updateLocation };

export default hot(module)(connect(mapStateToProps, mapDispatchToProps)(LoginCtrl));

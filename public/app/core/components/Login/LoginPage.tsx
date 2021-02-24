// Libraries
import React, { FC } from 'react';
import { css } from 'emotion';

// Components
import { UserSignup } from './UserSignup';
import { LoginServiceButtons } from './LoginServiceButtons';
import LoginCtrl from './LoginCtrl';
import { LoginForm } from './LoginForm';
import { ChangePassword } from '../ForgottenPassword/ChangePassword';
import { Branding } from 'app/core/components/Branding/Branding';
import { HorizontalGroup, LinkButton } from '@grafana/ui';
import { LoginLayout, InnerBox } from './LoginLayout';
import config from 'app/core/config';

const forgottenPasswordStyles = css`
  padding: 0;
  margin-top: 4px;
`;

export const LoginPage: FC = () => {
  document.title = Branding.AppTitle;
  return (
    <LoginLayout>
      <LoginCtrl>
        {({
          loginHint,
          passwordHint,
          ldapEnabled,
          authProxyEnabled,
          disableLoginForm,
          disableUserSignUp,
          login,
          oauthLogin,
          isLoggingIn,
          changePassword,
          skipPasswordChange,
          isChangingPassword,
          sessions,
          oauthName,
        }) => (
          <>
            {!isChangingPassword && (
              <InnerBox>
                {!disableLoginForm && (
                  <>
                    <LoginForm
                      onSubmit={login}
                      oauthLogin={oauthLogin}
                      loginHint={loginHint}
                      passwordHint={passwordHint}
                      isLoggingIn={isLoggingIn}
                      sessions={sessions}
                      oauthName={oauthName}
                    >
                      {!(ldapEnabled || authProxyEnabled) ? (
                        <HorizontalGroup justify="flex-end">
                          <LinkButton
                            className={forgottenPasswordStyles}
                            variant="link"
                            href={`${config.appSubUrl}/user/password/send-reset-email`}
                          >
                            Forgot your password?
                          </LinkButton>
                        </HorizontalGroup>
                      ) : (
                        <></>
                      )}
                    </LoginForm>
                  </>
                )}
                <LoginServiceButtons login={oauthLogin} />
                {!disableUserSignUp && <UserSignup />}
              </InnerBox>
            )}
            {isChangingPassword && (
              <InnerBox>
                <ChangePassword onSubmit={changePassword} onSkip={() => skipPasswordChange()} />
              </InnerBox>
            )}
          </>
        )}
      </LoginCtrl>
    </LoginLayout>
  );
};

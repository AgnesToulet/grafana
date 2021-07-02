// Libraries
import React, { PureComponent } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import AutoSizer from 'react-virtualized-auto-sizer';
import { dateTimeFormat } from '@grafana/data';

// Components
import Page from 'app/core/components/Page/Page';
import { Button, CodeEditor, HorizontalGroup, Icon, Modal } from '@grafana/ui';

// Actions & Selectors
import { getNavModel } from 'app/core/selectors/navModel';
import {
  loadDataSource,
  loadDataSourceHistory,
  loadDataSourceVersion,
  restoreDataSourceVersion,
} from './state/actions';
import { getDataSource } from './state/selectors';

// Types
import { DataSourceHistoryVersion, StoreState } from 'app/types';
import { GrafanaRouteComponentProps } from 'app/core/navigation/types';

export interface OwnProps extends GrafanaRouteComponentProps<{ uid: string }> {}

function mapStateToProps(state: StoreState, props: OwnProps) {
  const dataSourceId = props.match.params.uid;

  return {
    navModel: getNavModel(state.navIndex, `datasource-history-${dataSourceId}`),
    versions: state.dataSources.versions,
    version: state.dataSources.version,
    dataSource: getDataSource(state.dataSources, dataSourceId),
    isLoading: state.dataSources.isLoadingHistory,
    dataSourceId,
  };
}

const mapDispatchToProps = {
  restoreDataSourceVersion,
  loadDataSource,
  loadDataSourceHistory,
  loadDataSourceVersion,
};

interface State {
  showVersionModal: boolean;
}

const connector = connect(mapStateToProps, mapDispatchToProps);

export type Props = OwnProps & ConnectedProps<typeof connector>;

export class DataSourceHistory extends PureComponent<Props, State> {
  state = {
    showVersionModal: false,
  };

  async componentDidMount() {
    const { loadDataSource, dataSourceId } = this.props;
    await loadDataSource(dataSourceId);
    this.props.loadDataSourceHistory();
  }

  showVersionModal = (show: boolean) => () => {
    this.setState({ showVersionModal: show });
  };

  onRestore = (version: DataSourceHistoryVersion) => {
    const { dataSource, restoreDataSourceVersion } = this.props;
    restoreDataSourceVersion(dataSource, version);
  };

  onView = async (version: DataSourceHistoryVersion) => {
    const { loadDataSourceVersion } = this.props;
    await loadDataSourceVersion(version);
    this.showVersionModal(true)();
  };

  render() {
    const { versions, version, navModel, isLoading } = this.props;
    const { showVersionModal } = this.state;
    return (
      <Page navModel={navModel}>
        <Page.Contents isLoading={isLoading}>
          <table className="filter-table">
            <thead>
              <tr>
                <th />
                <th>Version</th>
                <th>Date</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {versions.map((version, index) => {
                return (
                  <tr key={`${version.version}-${index}`}>
                    <td className="width-1">
                      <Icon name="database" />
                    </td>
                    <td>{version.version}</td>
                    <td>{dateTimeFormat(parseInt(version.timestamp, 10) * 1000)}</td>
                    <td style={{ textAlign: 'right' }}>
                      <Button icon="history" variant="primary" size="sm" onClick={() => this.onRestore(version)}>
                        Restore
                      </Button>
                      <Button variant="secondary" size="sm" onClick={() => this.onView(version)}>
                        View
                      </Button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
          <VersionModal
            isOpen={showVersionModal}
            version={version}
            onDismiss={this.showVersionModal(false)}
            onRestore={this.onRestore}
          />
        </Page.Contents>
      </Page>
    );
  }
}

interface VersionModalProps {
  isOpen: boolean;
  version: DataSourceHistoryVersion;
  onDismiss(): void;
  onRestore(version: DataSourceHistoryVersion): void;
}

const VersionModal = ({ version, isOpen, onDismiss, onRestore }: VersionModalProps) => {
  return (
    <Modal title={`Version ${version.version}`} isOpen={isOpen} onDismiss={onDismiss}>
      <AutoSizer>
        {({ width, height }) => (
          <CodeEditor
            value={version.data || ''}
            language="json"
            width={width}
            height={height}
            showMiniMap={true}
            showLineNumbers={true}
          />
        )}
      </AutoSizer>
      <Modal.ButtonRow>
        <HorizontalGroup spacing="md" justify="center">
          <Button variant="primary" onClick={() => onRestore(version)}>
            Restore
          </Button>
          <Button variant="secondary" fill="outline" onClick={onDismiss}>
            Cancel
          </Button>
        </HorizontalGroup>
      </Modal.ButtonRow>
    </Modal>
  );
};

export default connector(DataSourceHistory);

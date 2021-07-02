// Libraries
import React, { PureComponent } from 'react';
import { connect, ConnectedProps } from 'react-redux';
import { dateTimeFormat } from '@grafana/data';

// Components
import Page from 'app/core/components/Page/Page';
import { Button, Icon } from '@grafana/ui';

// Actions & Selectors
import { getNavModel } from 'app/core/selectors/navModel';
import { loadDataSource, loadDataSourceHistory, restoreDataSourceVersion } from './state/actions';
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
    dataSource: getDataSource(state.dataSources, dataSourceId),
    isLoading: state.dataSources.isLoadingHistory,
    dataSourceId,
  };
}

const mapDispatchToProps = {
  restoreDataSourceVersion,
  loadDataSource,
  loadDataSourceHistory,
};

const connector = connect(mapStateToProps, mapDispatchToProps);

export type Props = OwnProps & ConnectedProps<typeof connector>;

export class DataSourceHistory extends PureComponent<Props> {
  async componentDidMount() {
    const { loadDataSource, dataSourceId } = this.props;
    await loadDataSource(dataSourceId);
    this.props.loadDataSourceHistory();
  }

  onRestore = (version: DataSourceHistoryVersion) => {
    const { dataSource, restoreDataSourceVersion } = this.props;
    restoreDataSourceVersion(dataSource.uid, version);
  };

  render() {
    const { versions, navModel, isLoading } = this.props;
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
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </Page.Contents>
      </Page>
    );
  }
}

export default connector(DataSourceHistory);

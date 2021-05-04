import defaults from 'lodash/defaults';

import React, { ChangeEvent, PureComponent } from 'react';
import { TextArea, Select, TagsInput, InlineFormLabel } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, SnowflakeOptions, SnowflakeQuery } from './types';

type Props = QueryEditorProps<DataSource, SnowflakeQuery, SnowflakeOptions>;

export class QueryEditor extends PureComponent<Props> {
  onQueryTextChange = (event: ChangeEvent<HTMLTextAreaElement>) => {
    const { onChange, query } = this.props;
    onChange({ ...query, queryText: event.target.value });
  };

  onQueryTypeChange = (value: SelectableValue<string>) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      queryType: value.value || 'table',
    });

    this.props.onRunQuery();
  };

  onUpdateColumnTypes = (columnKey: string, columns: string[]) => {
    const { onChange, query } = this.props;
    onChange({
      ...query,
      [columnKey]: columns,
    });

    this.props.onRunQuery();
  };

  options: Array<SelectableValue<string>> = [
    { label: 'Table', value: 'table' },
    { label: 'Time series', value: 'time series' },
  ];

  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText, queryType, timeColumns } = query;
    const selectedOption = this.options.find((options) => options.value === queryType) || this.options;

    return (
      <div>
        <div className="gf-form max-width-25" role="query-type-container">
          <InlineFormLabel width={5}>Query Type</InlineFormLabel>
          <Select
            width={20}
            allowCustomValue={false}
            isSearchable={false}
            onChange={this.onQueryTypeChange}
            options={this.options}
            value={selectedOption}
          />
        </div>
        <div className="gf-form">
          <TextArea
            css={null}
            style={{ height: 100 }}
            role="query-editor-input"
            value={queryText || ''}
            onBlur={() => this.props.onRunQuery()}
            onChange={this.onQueryTextChange}
            label="Query Text"
          />
        </div>
        {queryType === this.options[1].value && (
          <div className="gf-form">
            <div style={{ display: 'flex', flexDirection: 'column', marginRight: 15 }} role="time-column-selector">
              <InlineFormLabel>
                <div style={{ whiteSpace: 'nowrap' }}>Time formatted columns</div>
              </InlineFormLabel>
              <TagsInput
                onChange={(tags: string[]) => this.onUpdateColumnTypes('timeColumns', tags)}
                tags={timeColumns}
              />
            </div>
          </div>
        )}
      </div>
    );
  }
}

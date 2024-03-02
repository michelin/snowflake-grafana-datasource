import defaults from 'lodash/defaults';

import React, { PureComponent } from 'react';
import { Select, TagsInput, InlineFormLabel, CodeEditor, Field, Button } from '@grafana/ui';
import { QueryEditorProps, SelectableValue } from '@grafana/data';
import { DataSource } from './datasource';
import { defaultQuery, SnowflakeOptions, SnowflakeQuery } from './types';
import { format } from 'sql-formatter'

type Props = QueryEditorProps<DataSource, SnowflakeQuery, SnowflakeOptions>;

export class QueryEditor extends PureComponent<Props> {

  onQueryTextChange = (newQuery: string) => {
    const { onChange, query } = this.props;
    onChange({ ...query, queryText: newQuery });
  };

  onFormat = () => {
    try {
      let formatted = format(this.props.query.queryText || "", { 
        language: 'snowflake', 
        denseOperators: false, 
        keywordCase: 'upper',
      });
      // The formatter does not handle the $__ syntax correctly,
      // it adds a space after the method name before the bracket.
      // We fix that here.
      formatted = formatted.replace(/\$__(\w+)\s\(/g, '$__$1(');
      console.log("Formatted query", formatted)
      this.props.onChange({ ...this.props.query, queryText: formatted });
      console.log("Formatted query", formatted)
      this.props.onChange({ ...this.props.query, queryText: formatted });
    } catch (e) {
      console.log('Error formatting query', e);
    }

  }

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
          <InlineFormLabel width={10}>Query Type</InlineFormLabel>
          <Select
            width={20}
            allowCustomValue={false}
            isSearchable={false}
            onChange={this.onQueryTypeChange}
            options={this.options}
            value={selectedOption}
          />
        </div>
        <Field>
          <CodeEditor
            value={queryText || ''}
            onBlur={() => this.props.onRunQuery()}
            onChange={this.onQueryTextChange}
            language="sql"
            showLineNumbers={true}
            // width={'100%'}
            height={'200px'}
            showMiniMap={false}
            onSave={() => this.props.onRunQuery()}
          />
        </Field>
        <Field>
          <Button variant="secondary" onClick={this.onFormat}>Format</Button>
        </Field>
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

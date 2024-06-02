import defaults from 'lodash/defaults';

import React, { PureComponent } from 'react';
import { Select, TagsInput, InlineFormLabel, CodeEditor, Field, Button, RadioButtonGroup } from '@grafana/ui';
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
    { label: 'Time series', value: 'time series' }
  ];

  optionsFillMode: Array<SelectableValue<string>> = [
    { label: 'Don\'t fill', value: 'null' },
    { label: 'Keep previous value', value: 'previous' }
  ];

  onFillModeChange = (value: any) => {
    const {onChange, query} = this.props;
    onChange({
      ...query,
      fillMode: value || this.optionsFillMode[0].value,
    });
    this.props.onRunQuery();
  };


  render() {
    const query = defaults(this.props.query, defaultQuery);
    const { queryText, queryType, fillMode, timeColumns } = query;
    const selectedOption = this.options.find((options) => options.value === queryType) || this.options;
    const selectedFillMode = this.optionsFillMode.find((options) => options.value === fillMode)?.value || this.optionsFillMode[0].value;

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
          <div>
            <CodeEditor
              value={queryText || ''}
              onBlur={this.props.onRunQuery}
              onChange={this.onQueryTextChange}
              language="sql"
              showLineNumbers={true}
              height={'200px'}
              showMiniMap={false}
              onSave={this.props.onRunQuery}
            />
            <Button variant="secondary" icon="repeat" onClick={this.onFormat}>Format Query</Button>
          </div>
        </Field>
        {queryType === this.options[1].value && (
          <div>
            <Field label="Time series fill mode">
              <RadioButtonGroup value={selectedFillMode} options={this.optionsFillMode} onChange={this.onFillModeChange}/>
            </Field>
            <Field label="Time formatted columns">
              <TagsInput
                  width={40}
                  placeholder="Time series column name"
                  onChange={(tags: string[]) => this.onUpdateColumnTypes('timeColumns', tags)}
                  tags={timeColumns}
              />
            </Field>
          </div>
        )}
      </div>
    );
  }
}

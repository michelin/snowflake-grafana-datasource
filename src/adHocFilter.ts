import { getTable } from "ast";

export default class AdHocFilter {
  private _targetTable = '';

  setTargetTable(table: string) {
    console.log('adHocFilter: setTargetTable called')
    this._targetTable = table;
  }

  setTargetTableFromQuery(query: string) {
    console.log('adHocFilter: setTargetTableFromQuery called');
    this._targetTable = getTable(query);
    if (this._targetTable === '') {
      console.error('Failed to get table from adhoc query.');
      throw new Error('Failed to get table from adhoc query.');
    }
  }

  apply(sql: string, adHocFilters: AdHocVariableFilter[]): string {
    console.log('adHocFilter: apply called')
    if (sql === '' || !adHocFilters || adHocFilters.length === 0) {
      console.log('return \''+sql+'\' from adHocFilter empty check 1');
      return sql;
    }
    const filter = adHocFilters[0];
    if (filter.key.includes('.')) {
      this._targetTable = filter.key.split('.')[0];
    }
    if (this._targetTable === '' || !sql.match(new RegExp(`.*\\b${this._targetTable}\\b.*`, 'gi'))) {
      console.log('return \''+sql+'\' from adHocFilter emtpy check 2');
      return sql;
    }
    let filters = adHocFilters.map((f, i) => {
      const key = f.key.includes('.') ? f.key.split('.')[1] : f.key;
      const value = isNaN(Number(f.value)) ? `\\'${f.value}\\'` : Number(f.value);
      const condition = i !== adHocFilters.length - 1 ? (f.condition ? f.condition : 'AND') : '';
      return ` ${key} ${f.operator} ${value} ${condition}`;
    }).join('');
    sql = sql.replace(';', '');
    console.log('return \''+`${sql} settings additional_table_filters={'${this._targetTable}' : '${filters}'}`+'\' from adHocFilter');
    return `${sql} settings additional_table_filters={'${this._targetTable}' : '${filters}'}`;
  }
}

export type AdHocVariableFilter = {
  key: string;
  operator: string;
  value: string;
  condition: string;
};

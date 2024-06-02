import type { Configuration } from 'webpack';
import grafanaConfig from './.config/webpack/webpack.config';
import CopyWebpackPlugin from "copy-webpack-plugin";
import {hasReadme} from "./.config/webpack/utils";
import {forEach, mergeWith} from "lodash";

const config = async (env): Promise<Configuration> => {
    const baseConfig = await grafanaConfig(env);

    // merge with default config, but replace plugins section
    return mergeWith(baseConfig, {
        plugins: [
            new CopyWebpackPlugin({
                patterns: [
                    { from: hasReadme() ? 'README.md' : '../README.md', to: '.', force: true },
                    { from: 'plugin.json', to: '.' },
                    { from: '../LICENSE.txt', to: '.' },
                    { from: '../CHANGELOG.md', to: '.', force: true },
                    { from: '**/*.json', to: '.' },
                    { from: '**/*.svg', to: '.', noErrorOnMissing: true }, // Optional
                    { from: '**/*.png', to: '.', noErrorOnMissing: true }, // Optional
                    { from: '**/*.html', to: '.', noErrorOnMissing: true }, // Optional
                    { from: 'img/**/*', to: '.', noErrorOnMissing: true }, // Optional
                    { from: 'libs/**/*', to: '.', noErrorOnMissing: true }, // Optional
                    { from: 'static/**/*', to: '.', noErrorOnMissing: true }, // Optional
                ],
            }),
        ],
    }, (objValue, srcValue) => {
        forEach(objValue, (val, key) => {
           if (val.constructor.name === srcValue.constructor.name) {
                objValue[key] = srcValue[0]
           }
        });
    });
};

export default config;
const CopyWebpackPlugin = require('copy-webpack-plugin');
const webpack = require('webpack');
const path = require('path');
const http = require('http');

function resolve(dir) {
    return path.resolve(__dirname, dir)
}

const outputDir = process.env.WWW_OUT
    ? path.resolve(process.env.WWW_OUT)
    : resolve('../www');

module.exports = {
    outputDir,
    publicPath: '/',
    productionSourceMap: false,

    pages: {
        index: {
            entry: 'index.js',
            template: 'template.html',
            filename: 'index.html',
            title: 'LiveGBS',
            chunks: ['chunk-vendors', 'chunk-common', 'index']
        },
        login: {
            entry: 'login.js',
            template: 'template-login.html',
            filename: 'login.html',
            title: 'LiveGBS',
            chunks: ['chunk-vendors', 'chunk-common', 'login']
        },
        play: {
            entry: 'play.js',
            template: 'template-play.html',
            filename: 'play.html',
            title: 'LiveGBS',
            chunks: ['chunk-vendors', 'chunk-common', 'play']
        },
        playback: {
            entry: 'playback.js',
            template: 'template-playback.html',
            filename: 'playback.html',
            title: 'LiveGBS',
            chunks: ['chunk-vendors', 'chunk-common', 'playback']
        },
        map: {
            entry: 'map.js',
            template: 'template-map.html',
            filename: 'map.html',
            title: 'LiveGBS',
            chunks: ['chunk-vendors', 'chunk-common', 'map']
        },
        test: {
            entry: 'test.js',
            template: 'template-test.html',
            filename: 'test.html',
            title: 'LiveGBS',
            chunks: ['chunk-vendors', 'chunk-common', 'test']
        }
    },

    configureWebpack: {
        optimization: {
            splitChunks: {
                cacheGroups: {
                    default: false // 不自动提取共享chunk（消除长文件名）
                }
            }
        },
        externals: {
            jquery: 'window.$',
            'video.js': 'videojs',
            'flv.js': 'flvjs'
        },
        resolve: {
            extensions: ['.js', '.vue', '.json'],
            alias: {
                'vue$': 'vue/dist/vue.common.js',
                'assets': resolve('assets'),
                'components': resolve('components'),
                'elements': resolve('elements')
            }
        },
        plugins: [
            new webpack.ProvidePlugin({
                $: 'jquery',
                jQuery: 'jquery',
                "window.jQuery": 'jquery',
                "window.$": 'jquery'
            }),
            new CopyWebpackPlugin([
                { from: 'externals', to: outputDir },
                { from: 'node_modules/@liveqing/liveplayer/dist/component/liveplayer-lib.min.js', to: path.join(outputDir, 'js') },
                { from: 'node_modules/@liveqing/liveplayer/dist/component/liveplayer.swf', to: outputDir }
            ])
        ]
    },

    chainWebpack: config => {
        // Preserve whitespace in Vue templates (old vue-loader default)
        config.module
            .rule('vue')
            .use('vue-loader')
            .tap(options => {
                options.compilerOptions = {
                    whitespace: 'preserve'
                };
                return options;
            });

        // Remove all preload/prefetch plugins generically (not needed for multi-page)
        const removePlugins = [];
        config.plugins.store.forEach(plugin => {
            if (plugin.name && (plugin.name.startsWith('preload-') || plugin.name.startsWith('prefetch-'))) {
                removePlugins.push(plugin.name);
            }
        });
        removePlugins.forEach(name => config.plugins.delete(name));
    },

    css: {
        extract: {
            filename: 'css/[name].[contenthash:8].css',
            chunkFilename: 'css/[name].[contenthash:8].css'
        }
    },

    devServer: {
        host: '0.0.0.0',
        openPage: process.env.VUE_CLI_OPEN_PAGE || '',
        proxy: {
            '/': {
                target: 'http://127.0.0.1:10000',
                agent: new http.Agent({ keepAlive: true, maxSockets: 100 }),
            }
        }
    }
}

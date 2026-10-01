import Vue from 'vue';
import store from "./test-store";
import router from './test-router';
import hljs from 'highlight.js/lib/core';
import 'highlight.js/styles/atom-one-dark.css';
import highlightPlugin from '@highlightjs/vue-plugin';
import JsonViewer from 'vue-json-viewer';

hljs.registerLanguage('sip', function(hljs) {
    return {
        name: 'SIP',
        case_insensitive: false,
        keywords: {
            keyword: ['INVITE', 'ACK', 'BYE', 'CANCEL', 'OPTIONS', 'REGISTER',
                'MESSAGE', 'SUBSCRIBE', 'NOTIFY', 'PUBLISH', 'REFER', 'INFO'
            ],
            literal: ["Play", "Playback", "Download"],
        },
        contains: [
            {
                className: 'success',
                variants: [
                    {begin: /^SIP\/2\.0\s+(1\d{2}|2\d{2})\s+.+$/},
                    {begin: /<Result>OK<\/Result>/},
                    {begin: /<UpgradeResult>OK<\/UpgradeResult>/},
                ],
                relevance: 10
            },
            {
                className: 'error',
                variants: [
                    {begin: /^SIP\/2\.0\s+(4\d{2}|5\d{2})\s+.+$/},
                    {begin: /<Result>ERROR<\/Result>/},
                    {begin: /<UpgradeResult>ERROR<\/UpgradeResult>/},
                ],
                relevance: 15
            },
            {
                className: 'literal',
                begin: /\s(\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b)(?::\d+)?($|\s)/,
                relevance: 10
            },
            // hljs.QUOTE_STRING_MODE,      // 引用字符串
            // hljs.NUMBER_MODE,            // 数字（如端口号）
            {
                className: 'comment',
                begin: /^\s*\/\//,         // 注释行
                end: /$/
            }
        ]
    };
});

Vue.use(highlightPlugin)
Vue.use(JsonViewer)

new Vue({
    el: '#app',
    store,
    router,
    template: `
    <transition>
        <router-view></router-view>
    </transition>
    `
})
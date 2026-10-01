import Vue from "vue";
import Vuex from "vuex";
import moment from "moment";

Vue.use(Vuex);

const store = new Vuex.Store({
    state: {
        serverInfo: {},
        userInfo: null,
        work: true,
        ws: null,
        topic: "",
        url: "",
        msg: "",
        serial: localStorage["test_serial"]||"",
        code: localStorage["test_code"]||"",
        result: null,
        menus: [
            {
                path: '/catalog',
                title: "目录查询",
                icon: 'sitemap',
                roles: ['超级管理员'],
            }, {
                path: "/register",
                icon: "heartbeat",
                title: "注册心跳",
                roles: ['超级管理员'],
            }, {
                path: "/play",
                icon: "play",
                title: "实时点播",
                roles: ['超级管理员'],
            }, {
                path: "/ptz",
                icon: "arrows",
                title: "PTZ 遥控",
                roles: ['超级管理员'],
            }, {
                path: "/preset",
                icon: "flag",
                title: "预置位",
                roles: ['超级管理员'],
            }, {
                path: "/homeposition",
                icon: "home",
                title: "看守位",
                roles: ['超级管理员'],
            }, {
                path: "/ptzprecise",
                icon: "dot-circle-o",
                title: "精准 PTZ",
                roles: ['超级管理员'],
            }, {
                path: "/sdcard",
                icon: "hdd-o",
                title: "存储卡格式化",
                roles: ['超级管理员'],
            }, {
                path: "/targettrack",
                icon: "crosshairs",
                title: "目标跟踪",
                roles: ['超级管理员'],
            }, {
                path: "/iframe",
                icon: "clone",
                title: "强制关键帧",
                roles: ['超级管理员'],
            }, {
                path: "/playback",
                icon: "file-video-o",
                title: "录像回放/下载",
                roles: ['超级管理员'],
            }, {
                path: "/record",
                icon: "video-camera",
                title: "手动录像",
                roles: ['超级管理员'],
            }, {
                path: "/guard",
                icon: "exclamation-triangle",
                title: "布防/撤防",
                roles: ['超级管理员'],
            }, {
                path: "/alarmreset",
                icon: "bell",
                title: "报警复位",
                roles: ['超级管理员'],
            }, {
                path: "/infostatus",
                icon: "check-square-o",
                title: "状态信息查询",
                roles: ['超级管理员'],
            }, {
                path: "/teleboot",
                icon: "undo",
                title: "远程启动",
                roles: ['超级管理员'],
            }, {
                path: "/config",
                icon: "sliders",
                title: "配置参数",
                roles: ['超级管理员'],
            }, {
                path: "/cruisetrack",
                icon: "space-shuttle",
                title: "巡航轨迹查询",
                roles: ['超级管理员'],
            }, {
                path: "/subscribe",
                icon: "rss",
                title: "订阅/取消",
                roles: ['超级管理员'],
            }, {
                path: "/broadcast",
                icon: "microphone",
                title: "语音对讲",
                roles: ['超级管理员'],
            }, {
                path: "/upgrade",
                icon: "cogs",
                title: "软件升级",
                roles: ['超级管理员'],
            },
        ]
    },
    mutations: {
        updateUserInfo(state, userInfo) {
            state.userInfo = userInfo;
        },
        updateServerInfo(state, serverInfo) {
            state.serverInfo = serverInfo;
        },
        updateWork(state, work) {
            state.work = work;
        },
        updateWs(state, ws) {
            state.ws = ws;
        },
        updateTopic(state, topic) {
            state.topic = topic;
        },
        updateUrl(state, url) {
            state.url = url;
        },
        updateMsg(state, msg) {
            state.msg = msg;
        },
        updateSerial(state, serial) {
            localStorage["test_serial"] = serial;
            state.serial = serial
        },
        updateCode(state, code) {
            localStorage["test_code"] = code;
            state.code = code;
        },
        updateResult(state, result) {
            state.result = result;
        },
    },
    actions: {
        getUserInfo({ commit, state }, payload) {
            return new Promise((resolve, reject) => {
                $.get("/api/v1/userinfo", payload).then(msg => {
                    var userInfo = msg;
                    commit('updateUserInfo', userInfo);
                    resolve(userInfo);
                }).fail(() => {
                    resolve(null);
                })
            })
        },
        logout({ commit, state }) {
            return new Promise((resolve, reject) => {
                $.get("/api/v1/logout").then(data => {
                    commit('updateUserInfo', null);
                    resolve(null);
                }).fail(() => {
                    resolve(null);
                })
            })
        },
        getServerInfo({ commit }) {
            return new Promise((resolve, reject) => {
                $.ajax({
                    method: "GET",
                    global: false,
                    url: "/api/v1/getserverinfo",
                }).then(serverInfo => {
                    try {
                        if (serverInfo.ServerTime) {
                            var stime = moment(serverInfo.ServerTime, "YYYY-MM-DD HH:mm:ss");
                            serverInfo.DiffDuration = moment.duration(stime.diff(moment()));
                        }
                        commit('updateServerInfo', serverInfo);
                        resolve(serverInfo);
                        return
                    } catch (error) {
                        console.log(error);
                    }
                    resolve({});
                }).fail(() => {
                    resolve({});
                })
            })
        },
        check({ commit, state }) {
            return new Promise((resolve, reject) => {
                $.ajax({
                    method: "GET",
                    global: false,
                    url: "/api/v1/shark/work",
                }).then(ret => {
                    commit("updateWork", ret.IsWork && ret.CanWork);
                    resolve(ret.IsWork && ret.CanWork);
                }).fail(() => {
                    resolve(false);
                })
            })
        },
        connect({ commit, state }, payload = {}) {
            return new Promise((resolve, reject) => {
                if(state.ws) {
                    if(state.ws.readyState <= WebSocket.OPEN) {
                        resolve(state.ws);
                        return
                    }
                    state.ws.close();
                    commit("updateWs", null);
                }
                var protocol = "ws:";
                if(location.protocol.startsWith("https")) {
                    protocol = "wss:";
                }
                var serial = payload.serial||state.serial;
                if(!serial) {
                    resolve(null);
                    commit("updateWs", null);
                    return
                }
                var ws = new WebSocket(`${protocol}//${location.host}/api/v1/control/ws-shark/${serial}?topic=${payload.topic||state.topic||serial||'test'}`);
                var msg = "";
                ws.onopen = () => {
                    resolve(ws);
                    commit("updateWs", ws);
                    commit("updateUrl", "");
                }
                ws.onerror = err => {
                    resolve(null);
                    commit("updateWs", null);
                }
                ws.onclose = () => {
                    commit("updateWs", null);
                }
                ws.onmessage = evt => {
                    if(ws !== state.ws) return;
                    if(!msg && String(evt.data).trimEnd().endsWith(".pcap")) {
                        commit("updateUrl", `/api/v1/shark/download/${evt.data}`);
                        return
                    }
                    if(!msg && state.msg) {
                        msg = state.msg;
                    }
                    msg += evt.data;
                    // _.throttle
                    commit("updateMsg", msg);
                }
            })
        },
        disconnect({ commit, state }) {
            if(state.ws) {
                state.ws.close();
                commit("updateWs", null);

                if(!state.msg) {
                    commit("updateUrl", "");
                    // if(state.result && Array.isArray(state.result)) {
                    //     state.result.length = 0;
                    // }
                    // commit("updateResult", null);
                }
            }
        },
        play({ state }) {
            return new Promise(resolve => {
                $.ajax({
                    method: "POST",
                    url: "/api/v1/stream/start",
                    global: false,
                    data: {
                        serial: state.serial,
                        code: state.code,
                    }
                }).then((streamInfo = {}) => {
                    resolve(streamInfo);
                }).fail(() => {
                    resolve(null);
                })
            })
        },
        clear({ commit, state }) {
            if(state.ws) {
                state.ws.close();
                commit("updateWs", null);
            }
            commit("updateUrl", "");
            commit("updateMsg", "");
            if(state.result && Array.isArray(state.result)) {
                state.result.length = 0;
            }
            commit("updateResult", null);
        },
        sleep({ commit, state }, timeout = 3000) {
            return new Promise(resolve => {
                setTimeout(() => {
                    resolve("");
                }, timeout);
            })
        },
    }
})

export default store;
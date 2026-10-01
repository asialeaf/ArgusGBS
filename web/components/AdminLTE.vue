<template>
<div class="wrapper" v-if="$getQueryString('menu') != 'no'">
    <NaviBar :serverInfo="serverInfo" :userInfo="userInfo" @show-session-list="$refs['session-list-dlg'].show()"
		@show-stream-stats="$refs['stream-stats-dlg'].show()" @modify-password="$refs['password-modify-dlg'].show()">
	</NaviBar>
    <Sider :serverInfo="serverInfo" :userInfo="userInfo" :menus="menus"></Sider>
    <PasswordModifyDlg ref="password-modify-dlg" :serverInfo="serverInfo" :userInfo="userInfo"></PasswordModifyDlg>
    <SessionListDlg ref="session-list-dlg" :serverInfo="serverInfo" :userInfo="userInfo"></SessionListDlg>
	<StreamStatsDlg ref="stream-stats-dlg"></StreamStatsDlg>
    <div class="content-wrapper">
        <router-view @show-session-list="(filter, sms) => $refs['session-list-dlg'].show(filter, sms)"
			@show-stream-stats="(title, serial, code) => $refs['stream-stats-dlg'].show(title, serial, code)">
		</router-view>
    </div>
    <footer class="main-footer text-center">
        <span v-if="serverInfo.CopyrightText" v-html="serverInfo.CopyrightText"></span>
        <span v-else>
            Copyright © 2025 <a href="//www.liveqing.com" target="_blank">www.liveqing.com</a> All rights reserved.
        </span>
    </footer>
    <BackToTop class="hidden-xs"></BackToTop>
    <SidebarToggle class="hidden-xs"></SidebarToggle>
    <resize-observer @notify="handleResize"/>
</div>
<div class="wrapper" v-else>
    <div class="content-wrapper">
        <section class="content no-padding" style="overflow:hidden;">
            <router-view></router-view>
        </section>
    </div>
    <resize-observer @notify="handleResize"/>
</div>
</template>

<script>
import "font-awesome/css/font-awesome.css";
import "bootstrap/dist/css/bootstrap.css";
import "admin-lte/dist/css/skins/_all-skins.css";
import 'vue-resize/dist/vue-resize.css';
import "assets/styles/AdminLTE-custom.less";
import "assets/styles/custom.less";

import "bootstrap/dist/js/bootstrap.js";
import "admin-lte/dist/js/adminlte.js";

import 'assets/bootstrap-datepicker-1.7.1/css/bootstrap-datepicker3.less';
import 'assets/bootstrap-datepicker-1.7.1/js/bootstrap-datepicker.js';
import 'assets/bootstrap-datepicker-1.7.1/locales/bootstrap-datepicker.zh-CN.min.js';

import "leaflet/dist/leaflet.css";
import "@penggy/leaflet.fullscreen/Control.FullScreen.css";
import "@penggy/leaflet-contextmenu/dist/leaflet.contextmenu.css";

import { mapState } from "vuex";
import Vue from 'vue';
import moment from 'moment';
import Sider from 'components/Sider.vue';
import NaviBar from 'components/NaviBar.vue';
import BackToTop from 'components/BackToTop.vue';
import SidebarToggle from 'components/SidebarToggle.vue';
import PasswordModifyDlg from 'components/PasswordModifyDlg.vue';
import SessionListDlg from 'components/SessionListDlg.vue';
import StreamStatsDlg from 'components/StreamStatsDlg.vue';

import ElementUI from "element-ui";
import 'assets/styles/element-custom.scss';
import patchElementUI from "elements/patch.js";
Vue.use(ElementUI);
patchElementUI();

import VCharts from 'v-charts';
Vue.use(VCharts);

import VueClipboards from 'vue-clipboards';
Vue.use(VueClipboards);

import fullscreen from 'vue-fullscreen';
Vue.use(fullscreen);

import VueResize from 'vue-resize';
Vue.use(VueResize);

import VeeValidate, { Validator } from "vee-validate";
import zh_CN from "vee-validate/dist/locale/zh_CN";
Validator.localize("zh_CN", zh_CN);
Vue.use(VeeValidate, {
	locale: "zh_CN",
	// delay: 500,
	dictionary: {
		zh_CN: {
			messages: {
				required: field => `${field} 不能为空`,
				integer: field => `${field} 不是有效值`,
				confirmed: (field, targetField) => `${field} 和 ${targetField} 不匹配`,
				regex: field => `${field} 不符合要求格式`
			}
		}
	}
});
Vue.prototype.$updateQueryString = (uri, key, value) => {
    var re = new RegExp("([?&])" + key + "=.*?(&|$)", "i");
    var separator = uri.indexOf('?') !== -1 ? "&" : "?";
    if (uri.match(re)) {
        return uri.replace(re, '$1' + key + "=" + value + '$2');
    } else {
        return uri + separator + key + "=" + value;
    }
}
Vue.prototype.$getQueryString = (name, defVal = "") => {
	var reg = new RegExp("(^|&)" + name + "=([^&]*)(&|$)", "i");
	var r = window.location.search.substr(1).match(reg);
	if (r != null) {
		return decodeURIComponent(r[2]);
	}
	return defVal;
}
Vue.prototype.$iframe = (link, w, h) => {
    var _link = Vue.prototype.$updateQueryString(link, "aspect", `${w}x${h}`)
    return `<iframe src="${_link}" width="${w}" height="${h}" allowfullscreen allow="autoplay; fullscreen"></iframe>`
}
Vue.prototype.isMobile = () => {
	return videojs.browser.IS_IOS || videojs.browser.IS_ANDROID;
}
Vue.prototype.isIE = () => {
	return !!videojs.browser.IE_VERSION;
}
Vue.prototype.flvSupported = () => {
	return videojs.browser.IE_VERSION || (flvjs.getFeatureList() && flvjs.getFeatureList().mseLiveFlvPlayback && !videojs.browser.IS_IOS);
}
Vue.prototype.rtcSupported = () => {
    return !!window.RTCPeerConnection;
}
Vue.prototype.canTalk = () => {
	return location.protocol.indexOf("https") == 0 || location.hostname === 'localhost' || location.hostname === '127.0.0.1';
}
Vue.prototype.hasAnyRole = (serverInfo, userInfo, ...roles) => {
    roles = roles||[];
    if (serverInfo && serverInfo.APIAuth === false && !userInfo) {
        if(roles.length == 1 && roles[0] == '管理员') {
            return false;
        } else {
            return true;
        }
    }
    var userRoles = [];
    if (userInfo) {
        userRoles = userInfo.Roles || [];
    }
    var checked = false;
    for(var role of roles) {
        if (!role || userRoles.some(ur => (ur == role || ur == '超级管理员'))) {
            checked = true;
            break;
        }
    }
    return checked;
}
Vue.prototype.hasAllChannel = (serverInfo, userInfo) => {
    if (serverInfo && serverInfo.APIAuth === false && !userInfo) {
        return true;
    }
    if (userInfo) {
        return userInfo.HasAllChannel;
    }
    return false;
}
Vue.prototype.isDemoUser = (serverInfo, userInfo) => {
	return (serverInfo && userInfo && userInfo.Name === serverInfo.DemoUser);
}
Vue.prototype.formatDuration = (d, unit = 'seconds', lang)=> {
	if(!lang) {
		lang = document.getElementsByTagName('html')[0].getAttribute('lang') || 'zh-CN';
	}
	return moment.duration(d, unit).locale(lang).humanize();
}

import $ from "jquery";
import "@penggy/jquery.nicescroll";
$.ajaxSetup({ cache: false });
export default {
	data() {
		return {
			nice: null,
			optTimer: 0,
		}
	},
	mounted() {
		if(this.$getQueryString('menu') == 'no') {
            $('body').addClass("layout-top-nav");
		}
		$(document).ajaxError((evt, xhr, opts, ex) => {
			if (xhr.status == 401) {
				location.href = `/login?r=${encodeURIComponent(window.location.href)}`;
				return false;
			} else if (xhr.status) {
				let msg = xhr.responseText || "网络请求失败";
				if (xhr.status == 404) {
					msg = "请求服务不存在或已停止";
				} else if (xhr.status == 504) {
					msg = "Gateway Timeout";
				}
				try {
					msg = JSON.parse(msg)
				} catch (error) {}
				this.$message({
					type: 'error',
					message: msg
				})
			} else {
				console.log(`ajax error: ${xhr.status} ${xhr.responseText}`);
			}
		}).on("click", ".main-header .sidebar-toggle, .sidebar-toggle.hidden-xs", function () {
			setTimeout(() => {
				localStorage["sidebar-collapse"] = $("body").hasClass("sidebar-collapse") ? "sidebar-collapse" : "";
			}, 500)
		}).ready(() => {
			this.$nextTick(() => {
				$("body").layout("fix");
				this.fixHover();
				this.initNiceScroll();
			});
			if(this.serverInfo && this.serverInfo.OptTimeout > 0) {
				this.resetOptTimer();
				$(document).on("mousemove keydown scroll touchstart", this.resetOptTimer);
			}
			if(this.userInfo && this.userInfo.PwdModReq) {
				this.$refs['password-modify-dlg'].show();
				return;
			}
			if(this.userInfo && this.userInfo.PwdTipDays > 0) {
				this.$message({
					type: 'warning',
					showClose: true,
					duration: 5000,
					message: `你的密码超过${this.userInfo.PwdTipDays}天，建议修改密码`
				});
			}
			if(this.userInfo && (this.userInfo.LastLoginAt || this.userInfo.LastLoginIP)) {
				let msg = "登录成功";
				if(this.userInfo.LastLoginAt) {
					msg += `, 上次登录时间：${this.userInfo.LastLoginAt}`;
				}
				if(this.userInfo.LastLoginIP) {
					msg += `, 上次登录IP：${this.userInfo.LastLoginIP}`;
				}
				this.$message({
					type: 'success',
					showClose: true,
					duration: 5000,
					message: msg
				});
			}
		}).on('shown.bs.modal', () => {
			this.removeNiceScroll();
			$("body").css("overflow-y", "hidden");
		}).on('hidden.bs.modal', () => {
			$("body").css("overflow-y", "auto");
			this.initNiceScroll();
		});
		$("body").addClass(localStorage["sidebar-collapse"]);
	},
	components: {
		NaviBar, Sider, PasswordModifyDlg, SessionListDlg, StreamStatsDlg, BackToTop, SidebarToggle
	},
	computed: {
		...mapState([ "menus", "userInfo", "serverInfo" ]),
		// versionText(){
		// 	var text = "";
		// 	if(this.serverInfo){
		// 		text = this.serverInfo.Server || "";
		// 	}
		// 	var matchs = text.match(/LiveGBS\/(.+?)\s/i);
		// 	if(matchs) {
		// 		return matchs[1];
		// 	}
		// 	return ""
		// }
	},
	methods: {
		initNiceScroll() {
            if(!this.isIE() && !this.isMobile() && !this.nice) {
                this.nice = $('body').niceScroll({
                    zindex: 999999,
                    cursorwidth: "10px",
                    cursoropacitymax: 0.5,
                    preservenativescrolling: false,
                    enablekeyboard: false,
                });
                return true;
            }
            return false;
		},
		removeNiceScroll() {
            if (this.nice) {
                this.nice.remove();
                this.nice = null;
                return true;
            }
            return false;
		},
		fixHover() {
            if(videojs.browser.IS_IOS||videojs.browser.IS_ANDROID) {
                for(var sheetI = document.styleSheets.length - 1; sheetI >= 0; sheetI--) {
                    var sheet = document.styleSheets[sheetI];
                    if(sheet.cssRules) {
                        for(var ruleI = sheet.cssRules.length - 1; ruleI >= 0; ruleI--) {
                            var rule = sheet.cssRules[ruleI];
                            if(rule.selectorText) {
                                rule.selectorText = rule.selectorText.replace(":hover", ":active");
                                rule.selectorText = rule.selectorText.replace(":focus", ":active");
                            }
                        }
                    }
                }
            }
		},
		handleResize() {
			this.nice && this.nice.resize();
		},
		thisYear() {
			return moment().format('YYYY');
		},
		resetOptTimer() {
			if(this.optTimer) {
				clearTimeout(this.optTimer);
				this.optTimer = 0;
			}
			if(this.serverInfo && this.serverInfo.OptTimeout > 0) {
				this.optTimer = setTimeout(() => {
					let playing = false;
					for(const v of document.querySelectorAll("video")) {
						if(!v.paused && !v.ended && v.currentTime > 0) {
							playing = true;
							break;
						}
					}
					if(playing) {
						console.log(`${this.serverInfo.OptTimeout} 秒未操作，视频播放中`);
						this.resetOptTimer();
					} else {
						console.log(`${this.serverInfo.OptTimeout} 秒未操作，自动退出`);
						location.href = "/logout";
					}
				}, this.serverInfo.OptTimeout * 1000);
			}
		},
	},
	beforeRouteUpdate(to, from, next) {
		$('.modal').modal('hide'); // closes all active pop ups.
		next();
	},
	beforeDestroy() {
		$(document).off("mousemove keydown scroll touchstart", this.resetOptTimer);
	},
}
</script>

<style lang="less" scoped>
.content-wrapper, .right-side, .main-footer {
	transition: none;
}
</style>

<style lang="less">
.layout-top-nav {
	.box {
		margin-bottom: 0;
		border-top: 0;
	}
	.nav-tabs-custom {
		margin-bottom: 0;
	}
}
</style>

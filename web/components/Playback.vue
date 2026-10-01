<template>
<div class="wrapper" v-show="bReady">
	<div class="content-wrapper">
		<section class="content">
            <h4 class="text-primary text-center" v-if="name">{{name}}</h4>
            <div class="form-inline" autocomplete="off" spellcheck="false">
                <div class="form-group form-group-sm" v-if="type == 'device'">
                    <div class="btn-group btn-group-sm" role="group" v-if="speeder">
                        <button id="input-speed-dropdown" type="button" class="btn btn-default dropdown-toggle" data-toggle="dropdown">
                            {{speed}} 倍速 <span class="caret"></span>
                        </button>
                        <ul id="input-speed-menu" class="dropdown-menu">
                            <li :class="{active: speed == 0.5}"><a role="button" data-speed="0.5">0.5x</a></li>
                            <li :class="{active: speed == 1}"><a role="button" data-speed="1">1x</a></li>
                            <li :class="{active: speed == 2}"><a role="button" data-speed="2">2x</a></li>
                            <li :class="{active: speed == 4}"><a role="button" data-speed="4">4x</a></li>
                        </ul>
                    </div>
                    <span v-if="speeder">&nbsp;&nbsp;</span>
                    <button type="button" class="btn btn-default btn-sm" @click.prevent="showTimeRangeDlg" v-if="download && videos.length > 0"
                        v-loading.fullscreen.lock="bDownloading" element-loading-text="加载中...">
                        <i class="fa fa-download"></i> 下载
                    </button>
                </div>
                <div class="form-group form-group-sm" v-if="type == 'cloud'">
                    <button type="button" class="btn btn-default btn-sm" @click.prevent="showTimeRangeDlg" v-if="download && videos.length > 0"
                        v-loading.fullscreen.lock="bDownloading" element-loading-text="加载中...">
                        <i class="fa fa-download"></i> 下载
                    </button>
                </div>
                <div class="form-group form-group-sm pull-right" v-if="calendar">
                    <div class="input-group input-group-sm" v-if="serial && code && type == 'cloud' && calendar">
                        <CloudRecordDatePicker class="form-control" ref="datePicker"
                            :day="day" @update:day="updateDay" :serial="serial" :code="code" :token="token">
                        </CloudRecordDatePicker>
                        <div class="input-group-btn">
                            <button type="button" class="btn btn-sm btn-default" @click.prevent="showDatePicker">
                                <i class="fa fa-calendar"></i>
                            </button>
                        </div>
                    </div>
                    <div class="input-group input-group-sm" v-if="type == 'device' && calendar">
                        <DatePicker class="form-control" ref="datePicker"
                            @update:day="updateDay" :day="day">
                        </DatePicker>
                        <div class="input-group-btn">
                            <button type="button" class="btn btn-sm btn-default" @click.prevent="showDatePicker">
                                <i class="fa fa-calendar"></i>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
            <div class="clearfix"></div>
            <br>
            <LivePlayer ref="livePlayer" :videoUrl="videoUrl" :waterMark="osd" v-if="type == 'cloud'"
                :muted="muted" :controls="controls" :showCameraButton="snap" :showCustomButton="custom" :stretch="stretch" :autoplay="autoplay" :alt="alt"
				:debug="debug" :digitalZoom="digitalZoom" :dblclickFullscreen="dblclickFullscreen" :hideBigPlayButton="hideBigPlayButton !== undefined ? hideBigPlayButton : !!serverInfo.HideBigPlayButton"
                :currentTime="currentTime" @ended="onVideoEnd" @error="onVideoError" @timeupdate="onVideoTimeUpdate" @media_info="onMediaInfo"
                v-loading="loading" :loading.sync="loading" element-loading-text="加载中..." element-loading-background="#000">
            </LivePlayer>
            <LivePlayer ref="livePlayer" live hideLiveText :videoUrl="videoUrl" :waterMark="osd" v-if="type == 'device'"
                :muted="muted" :controls="controls" :showCameraButton="snap" :showCustomButton="custom" :stretch="stretch" :autoplay="autoplay" :alt="alt"
				:debug="debug" :digitalZoom="digitalZoom" :dblclickFullscreen="dblclickFullscreen" :hideBigPlayButton="hideBigPlayButton !== undefined ? hideBigPlayButton : !!serverInfo.HideBigPlayButton"
                @ended="onDeviceVideoEnd" @error="onDeviceVideoError" @timeupdate="onDeviceVideoTimeUpdate" @media_info="onMediaInfo"
				@pause="onDevicePause" @play="onDevicePlay"
                v-loading="loading" element-loading-text="加载中..." element-loading-background="#000">
            </LivePlayer>
            <div class="text-center" v-if="isDemoUser(serverInfo, userInfo) && !bOutHevcTip">
                <br>
                提示: 演示系统限制匿名登录播放时间, 若需测试长时间播放, 请<a target="_blank" href="//www.liveqing.com/docs/download/LiveGBS.html">下载使用</a>
            </div>
            <div class="text-center text-red" v-if="bOutHevcTip">
                <br>
                提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速
            </div>
            <br>
            <br>
            <CloudRecordTimeRule ref="timeRule" v-if="type == 'cloud'"
                :videos="videos" :start="start" :end="end" :minutesPerUnit="minutesPerUnit" @timeChange="onCloudTimeChange"  v-loading="loadingRecords">
            </CloudRecordTimeRule>
            <TimeRule ref="timeRule" v-if="type == 'device'"
                :videos="videos" :start="start" :end="end" :minutesPerUnit="minutesPerUnit" @timeChange="onDeviceTimeChange" v-loading="loadingRecords">
            </TimeRule>
            <br>
		</section>
	</div>
	<resize-observer @notify="handleResize"/>
	<TimeRangeDlg ref="timeRangeDlg" format="HHmmss" @submit="onTimeRangeChange"></TimeRangeDlg>
	<PlaybackDownloadDlg ref="playbackDownloadDlg" @download="onDownload"></PlaybackDownloadDlg>
</div>
</template>

<script>
import "font-awesome/css/font-awesome.css";
import "bootstrap/dist/css/bootstrap.css";
import "admin-lte/dist/css/skins/_all-skins.css";
import "vue-resize/dist/vue-resize.css";
import "assets/styles/AdminLTE-custom.less";
import "assets/styles/custom.less";

import "bootstrap/dist/js/bootstrap.js";
import "admin-lte/dist/js/adminlte.js";

import "assets/bootstrap-datepicker-1.7.1/css/bootstrap-datepicker3.less";
import "assets/bootstrap-datepicker-1.7.1/js/bootstrap-datepicker.js";
import "assets/bootstrap-datepicker-1.7.1/locales/bootstrap-datepicker.zh-CN.min.js";

import { mapState, mapActions } from "vuex";
import Vue from "vue";
import moment from "moment";

import ElementUI from "element-ui";
import "assets/styles/element-custom.scss";
import patchElementUI from "elements/patch.js";
Vue.use(ElementUI);
patchElementUI();

import VueResize from "vue-resize";
Vue.use(VueResize);

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
	var _link = Vue.prototype.$updateQueryString(link, "aspect", "fullscreen");
	return `<iframe src="${_link}" width="${w}" height="${h}" allowfullscreen allow="autoplay; fullscreen"></iframe>`;
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
Vue.prototype.hasAnyRole = (serverInfo, userInfo, ...roles) => {
    // toggle share on, no need login
    // if (serverInfo && serverInfo.APIAuth === false && !userInfo) {
    if (!userInfo) {
        return true;
    }
    var userRoles = [];
    if (userInfo) {
        userRoles = userInfo.Roles || [];
    }
    var checked = false;
    for(var role of (roles||[])) {
        if (!role || userRoles.some(ur => (ur == role || ur == '超级管理员'))) {
            checked = true;
            break;
        }
    }
    return checked;
}
Vue.prototype.isDemoUser = (serverInfo, userInfo) => {
	if (serverInfo && userInfo && userInfo.Name === serverInfo.DemoUser) return true;
	return false;
}

import CloudRecordDatePicker from 'components/CloudRecordDatePicker.vue';
import CloudRecordTimeRule from 'components/CloudRecordTimeRule.vue';
import DatePicker from 'components/DatePicker.vue';
import TimeRule from 'components/TimeRule.vue';
import TimeRangeDlg from 'components/TimeRangeDlg.vue';
import PlaybackDownloadDlg from 'components/PlaybackDownloadDlg.vue';
import LivePlayer from '@liveqing/liveplayer';
import $ from "jquery";
import "@penggy/jquery.nicescroll";
$.ajaxSetup({ cache: false });
export default {
	components: {
		LivePlayer, CloudRecordDatePicker, CloudRecordTimeRule, DatePicker, TimeRule, TimeRangeDlg, PlaybackDownloadDlg
	},
	data() {
		return {
			name: "",
			videoUrl: "",
			loading: false,
			loadingRecords: false,
			title: "",
			serial: "",
			code: "",
			type: "",
			protocol: "",
			alt: "无信号",
			osd: "",
			stretch: false,
			autoplay: true,
			controls: true,
			snap: undefined,
			hideBigPlayButton: undefined,
			custom: true,
			debug: false,
			digitalZoom: true,
			dblclickFullscreen: true,
			calendar: true,
			speeder: true,
			download: false,
			bDownloading: false,
			speed: 1,
			token: "",
			muted: true,
			otherParams: "",
			nice: null,
			day: moment().format('YYYYMMDD'),
			timerange: [
				moment(moment().format('YYYYMMDD'), "YYYYMMDD").startOf('hour').toDate(),
				moment(moment().format('YYYYMMDD'), "YYYYMMDD").startOf('hour').toDate()
			],
			currentTime: 0,
			speedTime: 0,
			speedAt: 0,
			minutes: 0,
			videos: [],
			video: null,
			streamID: "",
			records: [],
			recordListTimeout: 0,
			center: "",
			indistinct: "",
			switchTimer: 0, // switch to next video
			minutesPerUnit: 5,
			start: 0,
			end: 0,
			bReady: false,
			mediaInfo: null,
			bOutHevcTip: false,
			bPaused: false,
			resumeAt: "",
		};
	},
	beforeDestroy() {
		this.videoUrl = "";
		$(window).off("beforeunload", this.beforeUnload);
		$(document).off("click", "#input-speed-menu a", this.onSpeedClick);
	},
	created(){
		this.token = this.$getQueryString("token","");
		this.muted = this.$getQueryString("muted", "yes") == "yes";
		this.autoplay = this.$getQueryString("autoplay", "yes") == "yes";
		this.controls = this.$getQueryString("controls", "yes") == "yes";
		var snap = this.$getQueryString("snap", "");
		if(snap) {
			this.snap = (snap == "yes");
		}
		var hideBigPlayButton = this.$getQueryString("hide_big_play_button", "");
		if(hideBigPlayButton) {
			this.hideBigPlayButton = (hideBigPlayButton == "yes");
		}
		this.custom = this.$getQueryString("custom", "yes") == "yes";
		this.stretch = this.$getQueryString("stretch", "no") == "yes";
		this.debug = this.$getQueryString("debug") == "yes";
		this.digitalZoom = this.$getQueryString("digital_zoom", "yes") == "yes";
		this.dblclickFullscreen = this.$getQueryString("dblclick_fullscreen", "yes") == "yes";
		this.alt = this.$getQueryString("alt", "无信号");
		this.osd = this.$getQueryString("osd", "");
		this.speed = parseFloat(this.$getQueryString("speed", "1")) || 1;
	},
	async mounted() {
		var serverInfo = await this.getServerInfo();
		if (serverInfo) {
			document.title = serverInfo.LogoText || "LiveGBS";
		} else {
			serverInfo = {};
		}
		var payload = {};
		if(this.token) {
			payload["token"] = this.token;
		}
		await this.getUserInfo(payload);
		this.bReady = true;
		this.calendar = this.$getQueryString("calendar", "yes") == "yes";
		this.speeder = this.$getQueryString("speeder", "yes") == "yes";
		this.download = this.$getQueryString("download", "no") == "yes";
		this.serial = this.$getQueryString("serial", "");
		this.code = this.$getQueryString("code", "");
		this.center = this.$getQueryString("center", "");
		this.indistinct = this.$getQueryString("indistinct", "");
		this.protocol = this.$getQueryString("protocol", serverInfo.PreferStreamFmt||"");
		this.name = this.$getQueryString("name", "");
		var day = this.$getQueryString("day", moment().format('YYYYMMDD'));
		if(day) {
			if(moment(day, 'YYYYMMDD').isValid()) {
				this.day = day;
			} else {
				console.log(`invalid day[${day}]`);
			}
		}
		var minutes = this.$getQueryString("minutes", "");
		if(minutes) {
			this.minutes = parseInt(minutes) || 0;
			if(this.minutes < 0) {
				this.minutes = 0;
			}
			if(this.minutes > 1439) {
				this.minutes = 1439;
			}
		}
		var start = this.$getQueryString("start", "0");
		if(start) {
			this.start = parseInt(start) || 0;
			if(this.start < 0) {
				this.start = 0;
			}
			if(this.start > 1439) {
				this.start = 1439;
			}
			if(this.minutes < this.start) {
				this.minutes = this.start;
			}
		}
		var end = this.$getQueryString("end", "0");
		if(end) {
			this.end = parseInt(end) || 0;
			if(this.end < 0) {
				this.end = 0;
			}
			if(this.end > 1439) {
				this.end = 1439;
			}
			if(this.end && this.minutes > this.end) {
				this.minutes = this.end;
			}
		}
		var unit = this.$getQueryString("unit", this.$getQueryString("minutesPerUnit", serverInfo.MinutesPerUnit || "5"));
		if(unit) {
			this.minutesPerUnit = parseInt(unit) || 5;
			if (this.minutesPerUnit < 1 || this.minutesPerUnit > 60) {
				this.minutesPerUnit = 5;
			}
		}
		var recordListTimeout = this.$getQueryString("record_list_timeout", "")
		if(recordListTimeout) {
			this.recordListTimeout = parseInt(recordListTimeout) || 0;
			if(this.recordListTimeout < 0) {
				this.recordListTimeout = 0;
			}
		}
		this.type = this.$getQueryString("type", "device");
		this.otherParams = this.getOtherParams(["calendar", "speeder", "download", "speed", "minutes", "start", "end", "unit", "minutesPerUnit", "record_list_timeout", "autoplay",
			"controls", "snap", "custom", "stretch", "alt", "type", "serial", "code", "center", "indistinct", "protocol", "muted", "debug", "digital_zoom", "dblclick_fullscreen", "hide_big_play_button", "osd"]);
		$(document).ajaxError((evt, xhr, opts, ex) => {
			if (xhr.status == 401) {
				if (this.serverInfo.DemoUser) {
					location.href = `/login?r=${encodeURIComponent(location.href)}`;
				} else {
					console.log("登录认证过期");
				}
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
				console.log(msg);
			} else if(xhr) {
				console.log(`ajax error: ${xhr.status} ${xhr.responseText}`);
			}
		}).ready(() => {
			this.$nextTick(() => {
				$("body").layout("fix");
				this.fixHover();
				this.initNiceScroll();
			})
		}).on('shown.bs.modal', () => {
            this.removeNiceScroll();
		}).on('hidden.bs.modal', () => {
            this.initNiceScroll();
		}).on("click", "#input-speed-menu a", this.onSpeedClick);
		if(!this.minutes) {
			let mmt = moment();
			this.minutes = mmt.diff(mmt.clone().startOf('day'), 'minutes');
			this.minutes -= 10;
		}
		if (this.minutes < 0) this.minutes = 0;
		this.$nextTick(() => {
			var timeRule = this.$refs['timeRule'];
			timeRule && timeRule.clickMinute(this.minutes);
			if(this.type == 'cloud') {
				this.loadCloudVideos();
			} else if(this.type == 'device' && !this.calendar) {
				this.timerange[0] = moment(this.day, "YYYYMMDD").startOf('hour').toDate();
				this.timerange[1] = moment(this.day, "YYYYMMDD").startOf('hour').toDate();
				this.getDeviceRecords(true);
			}
		})
		$(window).on("beforeunload", this.beforeUnload);
	}, //-- mounted
	computed: {
		...mapState(["serverInfo", "userInfo"]),
		videoUrls() {
            return this.videos.map((val, index, array) => {
                return (val || {}).hls;
            });
		}
	},
	methods: {
		...mapActions(["getServerInfo", "getUserInfo"]),
		initNiceScroll() {
            if(!this.isIE() && !this.isMobile() && !this.nice) {
                this.nice = $('body').niceScroll({
                    zindex: 999999,
                    cursorwidth: "10px",
                    cursoropacitymax: 0.5,
                    preservenativescrolling: false,
                    enablekeyboard: false,
                });
            }
		},
		removeNiceScroll() {
            if (this.nice) {
                this.nice.remove();
                this.nice = null;
            }
		},
		fixHover() {
			if (videojs.browser.IS_IOS || videojs.browser.IS_ANDROID) {
				for (var sheetI = document.styleSheets.length - 1; sheetI >= 0; sheetI--) {
					var sheet = document.styleSheets[sheetI];
					if (sheet.cssRules) {
						for (var ruleI = sheet.cssRules.length - 1; ruleI >= 0; ruleI--) {
							var rule = sheet.cssRules[ruleI];
							if (rule.selectorText) {
								rule.selectorText = rule.selectorText.replace(":hover", ":active");
								rule.selectorText = rule.selectorText.replace(":focus", ":active");
							}
						}
					}
				}
			}
		},
		thisYear() {
			return moment().format("YYYY");
		},
		getOtherParams(without) {
			var url = location.search;
			var theRequest = "";
			if (url.indexOf("?") != -1) {
				var str = url.substr(1);
				var strs = str.split("&");
				for (var i = 0; i < strs.length; i++) {
					if (without.indexOf(strs[i].split("=")[0]) == -1) {
						if (theRequest == "") {
							theRequest = strs[i];
						} else {
							theRequest += "&" + strs[i];
						}
					}
				}
			}
			return theRequest;
		},
		handleResize() {
			this.nice && this.nice.resize();
		},
		updateDay(day) {
			// console.log(`update day[${day}] type[${this.type}]`);
			if(!moment(day, "YYYYMMDD").isValid()) return;
			this.day = day;
			if(this.type == 'cloud') {
				this.loadCloudVideos();
			} else if(this.type == 'device') {
				this.timerange[0] = moment(this.day, "YYYYMMDD").startOf('hour').toDate();
				this.timerange[1] = moment(this.day, "YYYYMMDD").startOf('hour').toDate();
				this.getDeviceRecords(true);
			}
		},
		showDatePicker() {
			$(this.$refs['datePicker'].$el).focus();
		},
		loadCloudVideos() {
			if(!this.serial || !this.code || this.type != 'cloud') return;
				this.loadingRecords = true;
				$.get("/api/v1/cloudrecord/querydaily", {
                    serial: this.serial,
                    code: this.code,
                    period: this.day,
                    token: this.token,
				}).then(data => {
                    // this.name = data.name || `${this.serial}:${this.code}`;
                    this.osd = this.osd || data.osd || "";
                    this.videos = data.list;
				}).always(() => {
                    this.loadingRecords = false;
				})
		},
		nextTimeRange() {
			var end = moment(this.day, "YYYYMMDD").endOf('day');
			var now = moment().startOf("second");
			if(end.isAfter(now, "second")) {
				end = now;
			}
			var r1 = moment(this.timerange[1]);
			if(r1.isSameOrAfter(end, "second")){
				return false;
			}
			var r2 = moment(this.timerange[1]).add(6, 'hours');
			if(r2.isAfter(end)) {
				r2 = end;
			}
			if(r2.clone().startOf("minute").isSameOrBefore(r1.clone().startOf("minute"), "second")) {
				return false;
			}
			console.log(r1.format("YY-MM-DD HH:mm:ss"), "~", r2.format("YY-MM-DD HH:mm:ss"), "loading...");
			this.timerange = [r1.toDate(), r2.toDate()];
			return true;
		},
		getDeviceRecords(refresh) {
			if(!this.serial || !this.code) return;
			if(refresh) {
				this.loadingRecords = true;
				this.records = [];
			}
			if(!this.nextTimeRange()){
				this.videos = this.records;
				this.loadingRecords = false;
				return
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/playback/recordlist",
				global: false,
				data: {
					timeout: this.recordListTimeout || this.serverInfo.RecordListTimeout || 5,
					serial: this.serial,
					code: this.code,
					center: this.center,
					indistinct: this.indistinct,
					starttime: moment(this.timerange[0]).format("YYYY-MM-DDTHH:mm:ss"),
					endtime: moment(this.timerange[1]).format("YYYY-MM-DDTHH:mm:ss"),
					token: this.token,
				}
			}).then(ret => {
				var items = ret.RecordList || [];
				this.records = this.records.concat(items.filter(item => {
					if(!item || !item.StartTime || !item.EndTime) {
						return false;
					}
					if(this.records.some(v => (v.StartTime === item.StartTime && v.EndTime === item.EndTime))) {
						console.log("repeat", item.StartTime, "~", item.EndTime);
						return false;
					}
					return true;
				}));
			}).always(() => {
				this.$nextTick(() => {
					this.getDeviceRecords(false);
				})
			});
		},
		async startPlayback(start, end) {
			if(!this.video) return;
			this.loading = true;
			await this.stopPlayback();
			$.post("/api/v1/playback/start", {
				serial: this.serial,
				code: this.code,
				starttime: start,
				endtime: end,
				token: this.token,
			}).then(streamInfo => {
				var protocol = "HLS";
				var videoUrl = streamInfo.HLS;
				if(this.flvSupported()) {
					videoUrl = this.isIE() ? streamInfo.WS_FLV : streamInfo.FLV;
					protocol = this.isIE() ? "WS_FLV" : "FLV";
				}
				switch(String(this.protocol).toUpperCase()) {
					case "WEBRTC":
						if(this.rtcSupported()) {
							videoUrl = streamInfo.WEBRTC;
							protocol = "WEBRTC";
						}
						break;
					case "WHEP":
						if(this.rtcSupported()) {
							videoUrl = streamInfo.WHEP;
							protocol = "WHEP";
						}
						break;
					case "FLV":
						if(this.flvSupported() && !this.isIE()) {
                            videoUrl = streamInfo.FLV;
                            protocol = "FLV";
						}
						break;
					case "WS_FLV":
					case "WS-FLV":
						if(this.flvSupported()) {
							videoUrl = streamInfo.WS_FLV;
							protocol = "WS_FLV";
						}
						break;
					case "HLS":
						videoUrl = streamInfo.HLS;
						protocol = "HLS";
						break;
					case "RTMP":
						videoUrl = streamInfo.RTMP;
						protocol = "RTMP";
						break;
				}
				this.streamID = streamInfo.StreamID;
				this.osd = this.osd || streamInfo.ChannelOSD || "";
				this.protocol = protocol;
				this.videoUrl = videoUrl;
				if(this.speed > 0 && this.speed != 1) {
					this.scaleOnStart();
				}
				this.resumeAt = start;
			}).always(() => {
				this.loading = false;
			});
		},
		stopPlayback() {
			return new Promise((resolve, reject) => {
                if(this.switchTimer) {
                    clearTimeout(this.switchTimer);
                    this.switchTimer = 0;
                }
                this.bOutHevcTip = false;
                this.videoUrl = "";
				this.resumeAt = "";
				this.currentTime = 0;
				this.speedTime = 0;
				this.speedAt = 0;
				this.bPaused = false;
                if(!this.streamID) {
                    resolve();
                    return
                }
                $.ajax({
                    method: "POST",
                    url: "/api/v1/playback/stop",
                    global: false,
                    data: {
                        streamid: this.streamID,
                        token: this.token,
                    }
                }).always(() => {
                    this.streamID = "";
                    // this.mediaInfo = null;
                    // this.bOutHevcTip = false;
                    resolve();
                })
			})
		},
		beforeUnload(event) {
			this.stopPlayback();
			// event.preventDefault();
			event.returnValue = '';
		},
		onCloudTimeChange(video, bSeek) {
			this.videoUrl = ((video || {}).hls || "");
			if(bSeek) {
				this.currentTime = (video || {}).currentTime || 0;
			} else {
				this.currentTime = 0;
			}
			this.bOutHevcTip = false;
		},
		onDeviceTimeChange(video) {
            if(video != this.video) {
                this.video = video;
                if(video) {
                    this.startPlayback(video.StartTime, video.EndTime);
                } else {
                    this.stopPlayback();
                }
            }
		},
		async onDeviceVideoError(e) {
			if(e == 'MediaError' && ((this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) || this.type == "cloud" || this.protocol == "HLS")) {
				if(flvjs.getFeatureList() && !flvjs.getFeatureList().nativeMP4H265Playback) {
					this.bOutHevcTip = true;
					console.log("提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速");
				}
			}
			if(!this.video) return;
			var startTime = moment(this.video.StartTime, "YYYY-MM-DDTHH:mm:ss");
			if(this.resumeAt && this.resumeAt > this.video.StartTime) {
				startTime = moment(this.resumeAt, "YYYY-MM-DDTHH:mm:ss");
			}
			if(startTime.format("YYYYMMDD") == this.day) {
				console.log("switch error", e, `at ${startTime.clone().add(this.currentTime, 'seconds').format("YYYY-MM-DDTHH:mm:ss")}`);
				if((e == "ended" || this.bPaused) && this.streamID && this.videoUrl) {
					var bResumed = await $.ajax({
						method: "POST",
						url: "/api/v1/playback/control",
						global: false,
						data: {
							streamid: this.streamID,
							command: "play"
						}
					}).then(() => {
						this.bPaused = false;
						if(this.protocol == "FLV" || this.protocol == "WS_FLV") {
							this.$refs["livePlayer"].flvReload();
						} else if(this.protocol == "WEBRTC" || this.protocol == "WHEP") {
							this.$refs["livePlayer"].rtcReload();
						} else {
							this.$refs["livePlayer"].reload();
						}
						this.resumeAt = startTime.clone().add(this.currentTime, 'seconds').format("YYYY-MM-DDTHH:mm:ss");
						console.log(`resume at ${this.resumeAt}`);
						this.currentTime = 0;
						this.speedTime = 0;
						this.speedAt = 0;
						return true;
					}, xhr => {
						xhr && console.log(`resume error: ${xhr.status} ${xhr.responseText}`);
						return false;
					});
					if(bResumed) return;
				}
				if(this.currentTime) {
					var endTime = moment(this.video.EndTime, "YYYY-MM-DDTHH:mm:ss");
					var playTime = startTime.clone().add(this.currentTime, 'seconds');
					var d = endTime.diff(playTime, 'seconds');
					if(d > 60) {
						this.startPlayback(playTime.format("YYYY-MM-DDTHH:mm:ss"), this.video.EndTime);
						return;
					}
				}
				for(var i = 1; i < 1440; i++) {
					var seekTime = startTime.clone().add(this.currentTime, 'seconds').add(i, 'minutes');
					if(!seekTime.isSame(startTime, "day")) {
						this.video = null;
						break;
					}
					var n = seekTime.diff(seekTime.clone().startOf('day'), 'minutes');
					var video = this.$refs['timeRule'].getVideo(n);
					if(video && video.StartTime != this.video.StartTime) {
						this.$refs['timeRule'].clickMinute(n);
						break;
					}
				}
			}
		},
		onDeviceVideoEnd() {
			this.mediaInfo = null;
		},
		onDeviceVideoTimeUpdate(t) {
			if(t < 1 || t < this.speedAt || !this.video) return;
			// this.bOutHevcTip = false;
			this.currentTime = this.speedTime + (t - this.speedAt) * this.speed;
			var startTime = moment(this.video.StartTime, 'YYYY-MM-DDTHH:mm:ss');
			if(this.resumeAt && this.resumeAt > this.video.StartTime) {
				startTime = moment(this.resumeAt, "YYYY-MM-DDTHH:mm:ss");
			}
			var playTime = startTime.clone().add(this.currentTime, 'seconds');
			var n = playTime.diff(playTime.clone().startOf('day'), 'minutes');
			this.$refs['timeRule'].clickMinute(n, false);
			var endTime = moment(this.video.EndTime, "YYYY-MM-DDTHH:mm:ss");
			var d = endTime.diff(playTime, 'seconds');
			if(d < 3 && !this.switchTimer && endTime.format("YYYYMMDD") == this.day) { // switch to next video
				console.log("switch timer", d);
				this.switchTimer = setTimeout(() => {
					for(var i = 1; i < 1440; i++) {
						var seekTime = endTime.clone().add(i, 'minutes');
						if(!seekTime.isSame(endTime, "day")) {
							this.video = null;
							break;
						}
						var n = seekTime.diff(seekTime.clone().startOf('day'), 'minutes');
						if(this.$refs['timeRule'].minuteActiveClass(n)) {
							this.$refs['timeRule'].clickMinute(n);
							break;
						}
					}
				}, 3000);
			}
		},
		onMediaInfo(mi) {
			this.mediaInfo = mi;
		},
		onVideoError(e) {
			if(e == 'MediaError' && ((this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) || this.type == "cloud" || this.protocol == "HLS")) {
				if(flvjs.getFeatureList() && !flvjs.getFeatureList().nativeMP4H265Playback) {
					this.bOutHevcTip = true;
					console.log("提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速");
				}
			}
		},
		onVideoEnd() {
            this.mediaInfo = null;
            var idx = this.videoUrls.indexOf(this.videoUrl);
            if (idx >= 0 && idx < this.videos.length - 1) {
                var nextVideo = this.videos[idx + 1];
                if (!nextVideo) return;
                var startTime = moment(nextVideo.startAt, 'YYYYMMDDHHmmss');
                var n = startTime.diff(startTime.clone().startOf('day'), 'minutes');
                this.$refs['timeRule'].clickMinute(n);
            }
		},
		onVideoTimeUpdate(t) {
            if (t < 1) return;
            // this.bOutHevcTip = false;
            var idx = this.videoUrls.indexOf(this.videoUrl);
            if (idx >= 0 && idx < this.videos.length) {
                var video = this.videos[idx];
                var startTime = moment(video.startAt, 'YYYYMMDDHHmmss').add(t, 'seconds');
                var n = startTime.diff(startTime.clone().startOf('day'), 'minutes');
                this.$refs['timeRule'].clickMinute(n, false);
            }
		},
		onSpeedClick(evt) {
			var speed = parseFloat($(evt.target).data("speed"));
			if(speed > 0 && this.speed != speed) {
				if(this.streamID) {
					this.scale(speed);
				} else {
					this.speed = speed;
				}
			}
		},
		onDevicePause() {
			if(!this.bPaused) {
				this.bPaused = true;
				if(this.streamID) {
					$.ajax({
						method: "POST",
						url: "/api/v1/playback/control",
						global: false,
						data: {
							streamid: this.streamID,
							command: "pause",
						}
					}).fail(xhr => {
						xhr && console.log(`pause stream[${this.streamID}] error: ${xhr.status} ${xhr.responseText}`);
					});
					this.$refs["livePlayer"].snap();
				}
			}
		},
		onDevicePlay() {
			if(this.bPaused) {
				this.bPaused = false;
				if(this.streamID) {
					$.ajax({
						method: "POST",
						url: "/api/v1/playback/control",
						global: false,
						data: {
							streamid: this.streamID,
							command: "play",
						}
					}).then(() => {
                        if(this.protocol == "WEBRTC" || this.protocol == "WHEP") {
                            this.$refs["livePlayer"].rtcReload();
                        }
                    }).fail(xhr => {
						xhr && console.log(`resume stream[${this.streamID}] error: ${xhr.status} ${xhr.responseText}`);
					});
				}
			}
		},
		scale(speed = 1) {
			if(!this.streamID) return;
			$.post("/api/v1/playback/control", {
				streamid: this.streamID,
				command: "scale",
				scale: speed,
				token: this.token,
			}).then(data => {
				this.speed = speed;
				this.speedTime = this.currentTime;
				this.speedAt = this.$refs["livePlayer"].getCurrentTime();
				this.$message({
                    type: "success",
                    message: `${speed} 倍速设置成功`,
				});
			}).fail(xhr => {
				xhr && this.$message({
					type: "error",
					message: xhr.responseText || "倍速设置失败",
				});
			})
		},
		scaleOnStart() {
			if(!this.streamID) return;
			$.ajax({
				method: "POST",
				url: "/api/v1/playback/control",
				global: false,
				data: {
					streamid: this.streamID,
					command: "scale",
					scale: this.speed,
				}
			})
		},
		async downloadPlayback(start, end) {
			this.bDownloading = true;
			await this.stopPlayback();
			$.post("/api/v1/playback/start", {
				serial: this.serial,
				code: this.code,
				starttime: start,
				endtime: end,
				download: true,
				token: this.token,
			}).then(streamInfo => {
				this.streamID = streamInfo.StreamID;
				this.$refs["playbackDownloadDlg"].download(start + " - " + end, streamInfo.StreamID);
			}).fail(xhr => {
				xhr && this.$message({
					type: "error",
					message: xhr.responseText || "下载失败",
				});
			}).always(() => {
				this.bDownloading = false;
			});
		},
		showTimeRangeDlg() {
			var start = moment().second(0).subtract(30, "minutes");
			var end = moment().second(0);
			if(!start.isSame(end, "day")) {
				start = end.clone().startOf("day");
			}
			var dlg = this.$refs['timeRangeDlg'];
			dlg.setDefault([start.toDate(), end.toDate()]);
			dlg.show();
		},
		onTimeRangeChange(v) {
			if(v && v.length == 2) {
				var start = `${this.day}${v[0]}`;
				var end = `${this.day}${v[1]}`;
				var t1 = moment(start, "YYYYMMDDHHmmss");
				var t2 = moment(end, "YYYYMMDDHHmmss");
				if(t1.isBefore(t2)) {
					var found = false;
					for(var i = t1.clone(); i.isSameOrBefore(t2); i = i.add(1, "minute")) {
                        var secs = i.unix();
                        secs = secs - (secs % ((this.serverInfo.MinutesPerUnit||5) * 60)); // floor to unit
                        var itext = moment.unix(secs).format("HH:mm");
                        if(this.$refs["timeRule"].activeMinutes[itext]) {
                            found = true;
                            break;
                        }
					}
					if(!found) {
                        this.$message({
                            type: "error",
                            message: `${t1.format("HH:mm:ss")} - ${t2.format("HH:mm:ss")} 没有录像`
                        })
                        return;
					}
					this.$refs['timeRangeDlg'].hide();
					if(this.type == "device") {
						this.downloadPlayback(t1.format("YYYY-MM-DDTHH:mm:ss"), t2.format("YYYY-MM-DDTHH:mm:ss"));
					} else if(this.type == "cloud") {
						this.bDownloading = true;
						var action = t1.seconds() % 10 != 0 || t2.seconds() % 10 != 0 ? "downloadex" : "download";
						var url = `/api/v1/cloudrecord/video/${action}/${this.serial}/${this.code}/${start}/${end}/video.mp4`;
						if(this.token) {
							url += `?token=${this.token}`;
						}
						$.get(url).then(ret => {
							window.location.href = ret;
						}).fail(xhr => {
							xhr && this.$message({
								type: "error",
								message: xhr.responseText || "下载失败",
							});
						}).always(() => {
                            this.bDownloading = false;
						})
					}
					return;
				}
			}
			this.$message({
                type: "error",
                message: `无效的时间段`
			})
		},
		onDownload() {
			this.streamID = "";
		},
	}
};
</script>

<style lang="less" scoped>
.main-header .navbar {
	/* Fix for IE */
	-webkit-transition: none;
	-o-transition: none;
	transition: none;
}

#input-speed-dropdown {
    color: #444 !important;
    background-color: #f4f4f4 !important;
    border-color: #ddd !important;
}

#input-speed-menu {
    min-width: unset;
}

.player-wrapper {
    margin: 0 auto;
    max-width: 700px;

    @media screen and(min-width: 1600px){
        max-width: 50%;
    }
}
</style>

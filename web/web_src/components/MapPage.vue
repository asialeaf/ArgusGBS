<template>
<div id="wrapper" v-loading="loading">
	<div id="map">
		<div class="form-group has-feedback search" v-show="bShowSearch">
            <input type="text" name="q" class="form-control" v-model.trim="q" autocomplete="off" placeholder="搜索" @mousedown.stop @touchstart.stop @dblclick.stop @contextmenu.stop>
            <span class="glyphicon glyphicon-search form-control-feedback text-gray"></span>
		</div>
	</div>
	<SimpleVideoDlg ref="videoDlg" id="video-dlg" :serverInfo="serverInfo" :userInfo="userInfo" :talk="talk" :protocol="protocol" :ptz="ptz" :token="token"
		:debug="debug" :digitalZoom="digitalZoom" :dblclickFullscreen="dblclickFullscreen" :hideBigPlayButton="hideBigPlayButton"
		:fluent="fluent" :stretch="stretch" :autoplay="autoplay" :controls="controls" :snap="snap" :custom="custom" :muted="muted">
	</SimpleVideoDlg>
	<el-dialog :title="msgDlgTitle" :visible.sync="bMsgDlg" :modal-append-to-body="false" :lock-scroll="false" width="450px" top="30vh">
		<span v-html="msgDlgText"></span>
		<span slot="footer" class="dialog-footer">
            <el-button type="primary" @click="copyMsgText" size="small">一键复制</el-button>
		</span>
	</el-dialog>
    <el-dialog title="设置" :visible.sync="bSettingsDlg" :modal-append-to-body="false" :lock-scroll="false" width="450px" top="30vh">
        <div class="form-horizontal">
            <div class="form-group">
                <label class="col-sm-4 control-label" for="input-refresh-interval">
                    自动刷新
                </label>
                <div class="col-sm-7">
                    <el-input-number id="input-refresh-interval" v-model.number="settingsForm.refreshInterval" :min="0" :step="5" size="small"></el-input-number>
                    <span class="text-gray" v-if="settingsForm.refreshInterval > 0">&nbsp;&nbsp;秒</span>
                    <span class="text-gray" v-else>&nbsp;&nbsp;不刷新</span>
                </div>
            </div>
            <div class="form-group">
                <label class="col-sm-4 control-label" for="input-tooltip">
                    显示浮窗
                </label>
                <div class="col-sm-7">
                    <select class="form-control" id="input-tooltip" name="tooltip" v-model.trim="settingsForm.tooltip">
                        <option value="">不显示</option>
                        <option value="permanent">始终显示</option>
                        <option value="hover">鼠标悬停显示</option>
                    </select>
                </div>
            </div>
            <div class="form-group" v-show="!!settingsForm.tooltip">
                <label class="col-sm-4 control-label" for="input-sticky">
                    粘性浮窗
                </label>
                <div class="col-sm-7">
                    <select class="form-control" id="input-sticky" name="sticky" v-model.trim="settingsForm.sticky">
                        <option :value="false">浮窗位置固定</option>
                        <option :value="true">浮窗跟随鼠标</option>
                    </select>
                </div>
            </div>
            <div class="form-group" v-show="!!settingsForm.tooltip">
                <label class="col-sm-4 control-label" for="input-snap-size">
                    快照大小
                </label>
                <div class="col-sm-7">
                    <el-radio-group v-model.trim="settingsForm.snapSize" size="mini">
                        <el-radio-button label="">无</el-radio-button>
                        <el-radio-button label="small">小</el-radio-button>
                        <el-radio-button label="medium">中</el-radio-button>
                        <el-radio-button label="large">大</el-radio-button>
                    </el-radio-group>
                </div>
            </div>
        </div>
        <!-- <span slot="footer" class="dialog-footer">
            <el-button type="primary" @click="bSettingsDlg = !bSettingsDlg" size="small">确定</el-button>
        </span> -->
    </el-dialog>
</div>
</template>

<script>
import "font-awesome/css/font-awesome.css";
import "bootstrap/dist/css/bootstrap.css";
import "admin-lte/dist/css/skins/_all-skins.css";
import "assets/styles/AdminLTE-custom.less";
import "assets/styles/custom.less";

import "bootstrap/dist/js/bootstrap.js";
import "leaflet/dist/leaflet.css";
import "@penggy/leaflet.fullscreen/Control.FullScreen.css";
import "@penggy/leaflet-contextmenu/dist/leaflet.contextmenu.css";
import "leaflet";
import "leaflet-polylinedecorator";
import "@penggy/leaflet.fullscreen";
import "@penggy/leaflet-contextmenu/dist/leaflet.contextmenu.js";

import { mapState, mapActions } from "vuex";
import Vue from "vue";

import ElementUI from "element-ui";
import 'assets/styles/element-custom.scss';
import patchElementUI from "elements/patch.js";
Vue.use(ElementUI);
patchElementUI();

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
	var _link = Vue.prototype.$updateQueryString(link, "aspect", "fullscreen")
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

import SimpleVideoDlg from "components/SimpleVideoDlg.vue";
import $ from "jquery";
import _ from "lodash";
$.ajaxSetup({ cache: false });

var camera_on = L.icon({
	iconUrl: "/images/camera-on.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var camera_off = L.icon({
	iconUrl: "/images/camera-off.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var camera_focus_on = L.icon({
	iconUrl: "/images/camera-red-on.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var camera_focus_off = L.icon({
	iconUrl: "/images/camera-red-off.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var dome_on = L.icon({
	iconUrl: "/images/dome-on.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var dome_off = L.icon({
	iconUrl: "/images/dome-off.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var dome_focus_on = L.icon({
	iconUrl: "/images/dome-red-on.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var dome_focus_off = L.icon({
	iconUrl: "/images/dome-red-off.png",
	iconSize: [30, 30],
	iconAnchor: [10, 30],
})
var defaultCenter = [31.82, 117.22];
var defaultZoom = 8;
var defaultMinZoom = 8;
var defaultMaxZoom = 9;
var defaultMarkerZoom = 9;
var defaultAttribution = "";
var defaultAttributionPrefix = "<a href='//www.liveqing.com' target='_blank'>LiveQing</a>";
export default {
	components: { SimpleVideoDlg },
	data() {
		return {
			q: "",
			serial: "",
			code: "",
			protocol: "",
			fluent: true,
			stretch: false,
			autoplay: true,
			controls: true,
			snap: undefined,
			hideBigPlayButton: undefined,
			custom: true,
			muted: "",
			debug: false,
			digitalZoom: true,
			dblclickFullscreen: true,
			ptz: true,
			talk: false,
			trace: false,
			search: true,
			bShowSearch: false,
			loading: false,
			token: "",
			otherParams: "",
			map: null,
			center: defaultCenter,
			zoom: defaultZoom,
			minZoom: defaultMinZoom,
			maxZoom: defaultMaxZoom,
			markerZoom: defaultMarkerZoom,
			attribution: defaultAttribution,
			attributionPrefix: defaultAttributionPrefix,
			channels: [],
			channelMarkers: [],
			channel: null,
			traceLine: null,
			traceLineDec: null,
			msgDlgTitle: "提示",
			msgDlgText: "",
			bMsgDlg: false,
            bSettingsDlg: false,
            settingsData: "",
            settingsForm: {
                tooltip: "hover",
                snapSize: "medium",
                sticky: false,
                refreshInterval: 0,
				focusDistance: 10000,
            },
            timer: 0,
		};
	},
	beforeDestroy() {
		if(this.timer) {
			clearTimeout(this.timer);
			this.timer = 0;
		}
		if(this.map) {
			this.map.remove();
			this.map = null;
		}
	},
	created(){
		this.muted = this.$getQueryString("muted", "");
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
		this.fluent = this.$getQueryString("fluent", "yes") == "yes";
		this.stretch = this.$getQueryString("stretch", "no") == "yes";
		this.debug = this.$getQueryString("debug", "no") == "yes";
		this.digitalZoom = this.$getQueryString("digital_zoom", "yes") == "yes";
		this.dblclickFullscreen = this.$getQueryString("dblclick_fullscreen", "yes") == "yes";
		this.token = this.$getQueryString("token", "");
	},
	async mounted() {
		var serverInfo = await this.getServerInfo();
		if (serverInfo) {
			document.title = serverInfo.LogoText || "LiveGBS";
		}
		await this.getUserInfo({
			token: this.token,
		});
		this.serial = this.$getQueryString("serial", "");
		this.code = this.$getQueryString("code", "");
		this.protocol = this.$getQueryString("protocol", serverInfo.PreferStreamFmt||"");
		this.ptz = this.$getQueryString("ptz", "yes") == "yes";
		this.talk = this.$getQueryString("talk", "no") == "yes";
		this.trace = this.$getQueryString("trace", "no") == "yes";
		this.search = this.$getQueryString("search", "yes") == "yes";
		this.otherParams = this.getOtherParams(["serial", "code", "protocol", "ptz", "talk", "trace", "search", "autoplay", "controls", "snap",
			"custom", "fluent", "stretch", "muted", "debug", "digital_zoom", "dblclick_fullscreen", "hide_big_play_button"]);
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
		})

        if(localStorage["livegbs_map_settings"]) {
            this.settingsData = localStorage["livegbs_map_settings"];
            try {
                let data = JSON.parse(this.settingsData);
                Object.assign(this.settingsForm, data);
            } catch (error) {
                console.log("load livegbs_map_settings error", error)
            }
        } else {
            this.settingsData = JSON.stringify(this.settingsForm);
        }
		var mapInfo = serverInfo.MapInfo || {};
		if(mapInfo.Center && mapInfo.Center.length >= 2) {
            this.center = [mapInfo.Center[1], mapInfo.Center[0]];
		}
		this.zoom = mapInfo.Zoom || defaultZoom;
		this.minZoom = mapInfo.MinZoom || defaultMinZoom;
		this.maxZoom = mapInfo.MaxZoom || defaultMaxZoom;
		this.markerZoom = mapInfo.MarkerZoom || defaultMarkerZoom;
		this.attribution = mapInfo.Attribution || defaultAttribution;
		this.attributionPrefix = mapInfo.AttributionPrefix || defaultAttributionPrefix;
		this.map = L.map("map", {
			fullscreenControl: true,
			fullscreenControlOptions: {
				fullscreenElement: this.$el,
			},
			contextmenu: true,
			contextmenuItems: [{
                text: "显示坐标",
                callback: this.showCoordinates,
			}],
			center: this.center,
			zoom: this.zoom,
			attributionControl: !!this.attribution,
		});
		L.tileLayer("/map/{z}/{x}/{y}.png", {
			minZoom: this.minZoom,
			maxZoom: this.maxZoom,
			attribution: this.attribution,
		}).addTo(this.map);
		if (this.map.attributionControl && this.attributionPrefix) {
			this.map.attributionControl.setPrefix(this.attributionPrefix);
		}

        if (this.map.zoomControl) {
            var settingsButton = L.DomUtil.create('a', 'leaflet-control-settings text-bold');
            settingsButton.innerHTML = '⚙️';
            settingsButton.title = '设置';
            settingsButton.style = 'font-size: 16px;';
            settingsButton.setAttribute('role', 'button');
            settingsButton.onclick = () => {
                this.bSettingsDlg = !this.bSettingsDlg;
            };
            this.map.zoomControl.getContainer().appendChild(settingsButton);
        }
		this.bShowSearch = this.search && this.zoom >= this.markerZoom;
		this.map.on('move', e => {
			if(this.map.getZoom() < this.markerZoom) {
				this.bShowSearch = false;
				if(this.channels.length > 0) {
					this.channels = [];
					this.refreshChannelMarkers();
				}
				return
			}
			this.bShowSearch = this.search;
			this.doDelaySearch();
		}).whenReady(() => {
			if(this.map.getZoom() >= this.markerZoom) {
				this.getChannels();
			}
			this.getChannelInfo();
		});
	}, //mounted
	computed: {
		...mapState(["serverInfo", "userInfo"]),
	},
	watch: {
		q: function(newVal, oldVal) {
            this.doDelaySearch();
		},
        bSettingsDlg: function(newVal, oldVal) {
            if(newVal) {
                this.settingsData = JSON.stringify(this.settingsForm);
            } else {
                let data = JSON.stringify(this.settingsForm);
                if(this.settingsData != data) {
                    localStorage["livegbs_map_settings"] = data;
                    this.refreshChannelMarkers(true);
                }
            }
        }
	},
	methods: {
		...mapActions(["getServerInfo", "getUserInfo"]),
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
		doDelaySearch: _.debounce(function(){
			this.getChannels();
		}, 800),
		getChannelInfo() {
			if(!this.serial || !this.code) return
			$.get("/api/v1/device/channelinfo", {
				serial: this.serial,
				code: this.code,
				token: this.token,
			}).then(channel => {
				var lat = channel.CustomLatitude || channel.Latitude;
				var lng = channel.CustomLongitude || channel.Longitude;
				if(!this.trace) {
					if(!this.channel) {
						this.map && this.map.setView([lat, lng], this.maxZoom);
					}
					this.channel = channel;
					return
				}
				$.get("/api/v1/device/positionlog", {
					serial: this.serial,
					code: this.code,
					token: this.token,
					sort: "CreatedAt",
					order: "asc",
				}).then(ret => {
					var latlngs = [];
					for(var log of ret.LogList) {
						latlngs.push([log.Latitude, log.Longitude])
					}
					if(this.traceLine) {
						this.traceLine.setLatLngs(latlngs);
						this.traceLineDec.setPaths(this.traceLine);
					} else {
						this.traceLine = L.polyline(latlngs, {weight: 8, color: 'red'});
						this.traceLineDec = L.polylineDecorator(this.traceLine, {
							patterns: [{
								offset: 50,
								repeat: 50,
								symbol: L.Symbol.arrowHead({
									pixelSize: 5,
									headAngle: 75,
									polygon: false,
									pathOptions: {
										stroke: true,
										weight: 2,
										color: 'white',
									}
								})
							}]
						});
					}
				}).always(() => {
					if(!this.channel) {
						this.map && this.map.setView([lat, lng], this.maxZoom);
					}
					this.channel = channel;
				})
			})
		},
		getChannels() {
			if(!this.map) return;
			if(this.map.getZoom() < this.markerZoom) {
				if(this.channels.length > 0) {
					this.channels = [];
					this.refreshChannelMarkers();
				}
				return
			}
			$.get("/api/v1/device/channellist", {
				channel_type: "device",
				bounds: this.map.getBounds().toBBoxString(),
				q: this.q,
				token: this.token,
			}).then(ret => {
				this.channels = ret.ChannelList || [];
				this.refreshChannelMarkers();
			})
		},
		refreshChannelMarkers(force = false) {
			if(!this.map) return;
			var channelMap = this.channels.reduce((pval, channel) => {
				var key = `${channel.DeviceID}_${channel.ID}`;
				pval[key] = channel;
				return pval;
			}, {})
			var moveFlag = false;
			this.channelMarkers = this.channelMarkers.filter(marker => {
				var channel = channelMap[marker.ID];
				if(channel && !force) {
					var lat = channel.CustomLatitude||channel.Latitude;
					var lng = channel.CustomLongitude||channel.Longitude;
					if(marker.Lat == lat && marker.Lng == lng && marker.Status == channel.Status
						&& (!this.settingsForm.tooltip || !this.settingsForm.snapSize || marker.SnapURL == channel.SnapURL)) {
						delete channelMap[marker.ID];
						if(this.serial == marker.Serial && this.code == marker.Code && this.map.getZoom() == this.maxZoom) {
							if(this.traceLine) this.traceLine.addTo(this.map);
							if(this.traceLineDec) this.traceLineDec.addTo(this.map);
						}
						return true;
					}
				}
				if(this.serial == marker.Serial && this.code == marker.Code) {
					if(this.map.getZoom() < this.maxZoom) {
						if(this.traceLine) this.traceLine.remove();
						if(this.traceLineDec) this.traceLineDec.remove();
					}
					if(this.trace && this.settingsForm.refreshInterval > 0 && (marker.Lat != lat || marker.Lng != lng)) {
						moveFlag = this.map.getCenter().distanceTo(marker.getLatLng()) <= this.settingsForm.focusDistance;
					}
				}
				marker.remove();
				return false;
			})
			for(var key in channelMap) {
				var channel = channelMap[key];
				var lat = channel.CustomLatitude||channel.Latitude;
				var lng = channel.CustomLongitude||channel.Longitude;
				var m = L.marker([lat, lng], {
					icon: channel.Status == "ON" ? camera_on : camera_off,
					title: this.settingsForm.tooltip ? "" : `${channel.CustomName || channel.Name || channel.ID} | ${channel.Status == "ON" ? "在线" : "离线"}`,
					contextmenu: true,
					contextmenuItems: [{
                        text: "设备信息",
                        callback: this.showChannelInfo,
					}],
				});
				m.ID = key;
				m.Lat = lat;
				m.Lng = lng;
				m.Serial = channel.DeviceID;
				m.Code = channel.ID;
				m.Status = channel.Status;
				m.ChannelInfo = channel;
				m.Title = channel.CustomName || channel.Name || channel.ID;
				m.PTZType = channel.CustomPTZType || channel.PTZType || 0;
				// if(m.PTZType === 1) {
					//   m.setIcon(channel.Status == "ON" ? dome_on : dome_off);
				// }
				m.SnapURL = channel.SnapURL || "";
				if(this.settingsForm.tooltip) {
                    let tooltip = `<span style="font-size:12px;">${m.Title}</span>`;
                    let snapWidth = 0, snapHeight = 0;
                    switch(this.settingsForm.snapSize) {
                        case "small":
                            snapWidth = 70;
                            snapHeight = 40;
                            break;
                        case "medium":
                            snapWidth = 144;
                            snapHeight = 81;
                            break;
                        case "large":
                            snapWidth = 288;
                            snapHeight = 162;
                            break;
                    }
                    if(snapWidth > 0 && snapHeight > 0) {
                        if(tooltip) {
                            tooltip += "<br>"
                        }
                        tooltip += `<img src="${m.SnapURL}" width="${snapWidth}" height="${snapHeight}" onerror="this.src='/images/default_snap.png';"/>`;
                    }
					m.bindTooltip(tooltip, {
						direction: 'top',
						offset: [0, -25],
						opacity: 1,
						className: "tooltip-snap",
						sticky: this.settingsForm.sticky,
						permanent: this.settingsForm.tooltip == 'permanent',
					})
				}
				if(channel.Status == "ON") {
					m.on('click', this.onChannelMarkerClick);
				} else {
					m.on('dblclick', this.onChannelMarkerClick);
				}
				m.addTo(this.map);
				this.channelMarkers.push(m);
				if(this.serial == channel.DeviceID && this.code == channel.ID) {
					m.setIcon(channel.Status == "ON" ? camera_focus_on : camera_focus_off);
					// if(m.PTZType === 1) {
					//   m.setIcon(channel.Status == "ON" ? dome_focus_on : dome_focus_off);
					// }
					m.setZIndexOffset(999);
					if(this.traceLine) this.traceLine.addTo(this.map);
					if(this.traceLineDec) this.traceLineDec.addTo(this.map);
					if(moveFlag) this.map.setView([lat, lng], this.map.getZoom());
				}
			}
            if(this.timer) {
                clearTimeout(this.timer);
                this.timer = 0;
            }
            if(this.settingsForm.refreshInterval > 0) {
                this.timer = setTimeout(() => {
                    this.getChannels();
					this.getChannelInfo();
                }, this.settingsForm.refreshInterval * 1000);
            }
		},
		onChannelMarkerClick(e) {
			var marker = e.sourceTarget;
			if(!marker) return;
			var channel = marker.ChannelInfo;
			if(!channel) return;
			this.loading = true;
			$.ajax({
				method: "POST",
				url: "/api/v1/stream/start",
				global: false,
				data: {
					serial: channel.DeviceID,
					code: channel.ID,
					token: this.token,
				}
			}).then(streamInfo => {
				this.$refs["videoDlg"].play(channel.DeviceID, channel.ID, streamInfo);
			}).fail(xhr => {
				var msg = "加载视频失败";
				if(channel.Status != "ON") {
					msg = "设备离线";
				} else if(xhr) {
					msg = `${xhr.status} ${xhr.responseText}`;
				}
				this.$refs["videoDlg"].play(channel.DeviceID, channel.ID, {
					ChannelName: channel.CustomName || channel.Name,
					Message: msg,
				})
			}).always(() => {
				this.loading = false;
			})
		},
		showCoordinates(e) {
            this.showMsgDlg("坐标(经,纬)", `${parseFloat(e.latlng.lng.toFixed(5))},${parseFloat(e.latlng.lat.toFixed(5))}`);
		},
		showChannelInfo(e) {
            var m = e.relatedTarget;
            var msg =  `编号: ${m.ID}<br>`;
            msg += `名称: ${m.Title}<br>`;
            msg += `经纬: ${parseFloat(m.Lng.toFixed(5))},${parseFloat(m.Lat.toFixed(5))}`;
            this.showMsgDlg("设备信息", msg);
		},
		showMsgDlg(title, text) {
            this.msgDlgTitle = title;
            this.msgDlgText = text;
            this.bMsgDlg = true;
		},
		copyMsgText() {
            this.copy(this.msgDlgText.replaceAll("<br>", "\r\n"));
            this.bMsgDlg = false;
		},
		centerHere(e) {
            this.map.panTo(e.latlng);
		},
		async copy(text) {
			try {
				let textArea = document.createElement("textarea");
				textArea.value = text;
				textArea.style.position = "absolute";
				textArea.style.opacity = 0;
				textArea.style.left = "-999999px";
				textArea.style.top = "-999999px";
				this.$el.appendChild(textArea);
				textArea.focus();
				textArea.select();
				await new Promise((res, rej) => {
					document.execCommand('copy') ? res() : rej();
					textArea.remove();
				});
				this.$message.success('复制成功！');
			} catch (err) {
				try {
					await navigator.clipboard.writeText(text);
					this.$message.success('复制成功！！');
				} catch (errr) {
					console.log(err, errr);
					this.$message.error('复制失败', err, errr);
				}
			}
		},
	}, // methods
};
</script>

<style lang="less">
html,body {
	height: 100%;
	margin: 0;
	padding: 0;
}
</style>
<style lang="less" scoped>
#wrapper {
	width: 100%;
	height: 100%;

	#map {
		width: 100%;
		height: 100%;

		.form-group.search {
            position: absolute;
            z-index: 1000;
            top: 15px;
            right: 15px;
            width: 200px;

			.glyphicon-search::before {
				font-size: 12px;
			}
		}
	}
}
</style>

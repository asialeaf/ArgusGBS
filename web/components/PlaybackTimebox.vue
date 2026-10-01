<template>
<div class="">
    <div class="box box-primary">
        <div class="box-header">
            <h4 class="text-primary text-center">设备录像时间轴({{devid}}-{{channel}})</h4>
        </div>
        <div class="box-body">
            <div class="form-inline" autocomplete="off" spellcheck="false">
                <div class="form-group form-group-sm">
                    <button type="button" class="btn btn-sm btn-primary" @click.prevent="$router.go(-1)" v-if="$getQueryString('menu') != 'no'">
                        <i class="fa fa-chevron-left"></i> 返回
                    </button>
                    <span v-if="shareURL">&nbsp;&nbsp;</span>
                    <a role="button" class="btn btn-default btn-sm" :href="shareURL" target="_blank" v-if="shareURL">
                        <i class="fa fa-share"></i> 分享页
                    </a>
                    <span>&nbsp;&nbsp;</span>
                    <div class="btn-group btn-group-sm" role="group">
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
                    <span v-if="!isDemoUser(serverInfo, userInfo) && videos.length > 0">&nbsp;&nbsp;</span>
                    <button type="button" class="btn btn-default btn-sm" @click.prevent="showTimeRangeDlg" v-if="!isDemoUser(serverInfo, userInfo) && videos.length > 0"
                        v-loading.fullscreen.lock="bDownloading" element-loading-text="加载中...">
                        <i class="fa fa-download"></i> 下载
                    </button>
                </div>
                <div class="form-group pull-right">
                    <div class="input-group input-group-sm">
                        <DatePicker class="form-control" @update:day="updateDay" :day="day" ref="datePicker"></DatePicker>
                        <div class="input-group-btn">
                            <button type="button" class="btn btn-sm btn-default" @click.prevent="showDatePicker">
                                <i class="fa fa-calendar"></i>
                            </button>
                            <button @click.prevent="toListView" class="btn btn-default btn-sm">
                                <i class="fa fa-hand-o-right"></i> 列表视图
                            </button>
                        </div>
                    </div>
                </div>
            </div>
            <br>
            <div class="clearfix"></div>
            <LivePlayer ref="livePlayer" :videoUrl="videoUrl" :muted="!bAudioEnable" :waterMark="osd" :hideBigPlayButton="!!serverInfo.HideBigPlayButton" live hideLiveText digitalZoom
                @media_info="onMediaInfo" @ended="onEnded" @error="onVideoError" @timeupdate="onVideoTimeUpdate" @pause="onPause" @play="onPlay"
                v-loading="videoLoading" element-loading-text="加载中..." element-loading-background="#000">
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
            <TimeRule :videos="videos" @timeChange="onTimeChange" ref="timeRule" v-loading="loading" :minutesPerUnit="serverInfo.MinutesPerUnit||5"></TimeRule>
            <div class="clearfix"></div>
            <br>
        </div>
    </div>
    <TimeRangeDlg ref="timeRangeDlg" format="HHmmss" @submit="onTimeRangeChange"></TimeRangeDlg>
    <PlaybackDownloadDlg ref="playbackDownloadDlg" @download="onDownload"></PlaybackDownloadDlg>
</div>
</template>

<script>
import _ from 'lodash'
import moment from 'moment'
import DatePicker from 'components/DatePicker.vue'
import TimeRule from 'components/TimeRule.vue'
import TimeRangeDlg from 'components/TimeRangeDlg.vue'
import PlaybackDownloadDlg from 'components/PlaybackDownloadDlg.vue'
import LivePlayer from '@liveqing/liveplayer'
import { mapState } from 'vuex'

export default {
	props: {
		devid: {
			type: String,
			default: ""
		},
		channel: {
			type: String,
			default: ""
		},
		mode: {
			type: String,
			default: "timebox"
		},
		day: {
			type: String,
			default: () => moment().format("YYYYMMDD")
		}
	},
	data() {
		return {
			timerange: [
				moment(this.day, "YYYYMMDD").startOf('hour').toDate(),
				moment(this.day, "YYYYMMDD").startOf('hour').toDate()
			],
			videoLoading: false,
			loading: false,
			center: "",
			indistinct: "",
			records: [],
			videos: [],
			video: null,
			protocol: "",
			videoUrl: "",
			streamID: "",
			osd: "",
			currentTime: 0,
			speedTime: 0,
			speedAt: 0,
			touchTimer: 0,
			switchTimer: 0,
			speed: 1,
			mediaInfo: null,
			bOutHevcTip: false,
			bDownloading: false,
			bShared: false,
			bAudioEnable: false,
			bPaused: false,
			resumeAt: "",
		};
	},
	computed: {
		...mapState(['userInfo', 'serverInfo']),
		shareURL() {
			if(this.bShared || this.serverInfo.APIAuth === false || this.serverInfo.DemoUser) {
				return `/playback.html?type=device&serial=${this.devid}&code=${this.channel}&day=${this.day}`;
			}
			return "";
		}
	},
	components: {
		DatePicker, LivePlayer, TimeRule, TimeRangeDlg, PlaybackDownloadDlg
	},
	methods: {
		ready() {
			this.$watch('day', function(newVal, oldVal) {
				this.timerange = [
					moment(this.day, "YYYYMMDD").startOf('hour').toDate(),
					moment(this.day, "YYYYMMDD").startOf('hour').toDate()
				]
			});
			this.$watch('center', function(newVal, oldVal) {
				this.timerange = [
					moment(this.day, "YYYYMMDD").startOf('hour').toDate(),
					moment(this.day, "YYYYMMDD").startOf('hour').toDate()
				]
			});
			this.$watch('indistinct', function(newVal, oldVal) {
				this.timerange = [
					moment(this.day, "YYYYMMDD").startOf('hour').toDate(),
					moment(this.day, "YYYYMMDD").startOf('hour').toDate()
				]
			});
			this.$watch('video', function(newVal, oldVal) {
				if(newVal && newVal != oldVal) {
					this.startPlayback(newVal.StartTime, newVal.EndTime);
				} else {
					this.stopPlayback();
				}
			});

			let mmt = moment();
			let n = mmt.diff(mmt.clone().startOf('day'), 'minutes');
			n -= 10;
			if(n < 0) n = 0;
			this.$refs.timeRule.clickMinute(n);
			console.log(this.devid, this.channel, this.day);
			this.getRecords(true);
			$(window).on("beforeunload", this.beforeUnload);
		},
		keyDown(e) {
			if(e.keyCode == 27) {
				this.$el.querySelector('.fa-chevron-left').click();
			}
		},
		isMobile() {
			return videojs.browser.IS_IOS || videojs.browser.IS_ANDROID;
		},
		showDatePicker() {
			$(this.$refs.datePicker.$el).focus();
		},
		updateDay(day) {
			this.$nextTick(() => {
				var rq = {};
				if(this.center) {
					rq.center = this.center;
				}
				if(this.indistinct) {
					rq.indistinct = this.indistinct;
				}
				this.$router.replace({
					path: `/devices/playback/${this.mode}/${this.devid}/${this.channel}/${day}`,
					query: Object.assign({}, this.$route.query, rq),
				});
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
		getRecords(refresh) {
			if(refresh) {
				this.loading = true;
				this.records = [];
			}
			if(!this.nextTimeRange()){
				this.videos = this.records;
				this.$nextTick(() => {
					this.loading = false;
				})
				return
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/playback/recordlist",
				global: false,
				data: {
					timeout: this.serverInfo.RecordListTimeout || 5,
					serial: this.devid,
					code: this.channel,
					center: this.center,
					indistinct: this.indistinct,
					starttime: moment(this.timerange[0]).format("YYYY-MM-DDTHH:mm:ss"),
					endtime: moment(this.timerange[1]).format("YYYY-MM-DDTHH:mm:ss")
				}
			}).then(ret => {
				this.bShared = !!ret.Shared;
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
					this.getRecords(false);
				})
			});
		},
		stopPlayback() {
			return new Promise((resolve, reject) => {
                if(this.switchTimer) {
                    clearTimeout(this.switchTimer);
                    this.switchTimer = 0;
                }
                this.bOutHevcTip = false;
				this.bAudioEnable = false;
                this.videoUrl = "";
                this.osd = "";
				this.resumeAt = "";
				this.currentTime = 0;
				this.speedTime = 0;
				this.speedAt = 0;
				this.bPaused = false;
                if(this.touchTimer) {
                    clearInterval(this.touchTimer);
                    this.touchTimer = 0;
                }
                if(!this.streamID) {
                    resolve();
                    return
                }
                $.ajax({
                    method: "POST",
                    url: "/api/v1/playback/stop",
                    global: false,
                    data: {
                        streamid: this.streamID
                    }
                }).always(() => {
					resolve();
                })
                this.streamID = "";
                // this.mediaInfo = null;
                // this.bOutHevcTip = false;
			})
		},
		beforeUnload(event) {
			this.stopPlayback();
			// event.preventDefault();
			event.returnValue = '';
		},
		async startPlayback(start, end) {
			this.videoLoading = true;
			await this.stopPlayback();
			$.post("/api/v1/playback/start", {
				serial: this.devid,
				code: this.channel,
				starttime: start,
				endtime: end,
			}).then(streamInfo => {
				console.log(`playback start ${start} ~ ${end}`);
				var protocol = "HLS";
				var videoUrl = streamInfo.HLS;
				if(this.flvSupported()) {
					videoUrl = this.isIE() ? streamInfo.WS_FLV : streamInfo.FLV;
					protocol = this.isIE() ? "WS_FLV" : "FLV";
				}
				switch(String(this.serverInfo.PreferStreamFmt).toUpperCase()) {
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
							protocol = "WS_FLV"
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
				this.bAudioEnable = !!streamInfo.AudioEnable;
				this.osd = streamInfo.ChannelOSD || "";
				this.protocol = protocol;
				this.videoUrl = videoUrl;
				if(this.speed > 0 && this.speed != 1) {
					this.scaleOnStart();
				}
				this.resumeAt = start;
			}).always(() => {
				this.videoLoading = false;
			});
		},
		async downloadPlayback(start, end) {
			this.bDownloading = true;
			await this.stopPlayback();
			$.post("/api/v1/playback/start", {
				serial: this.devid,
				code: this.channel,
				starttime: start,
				endtime: end,
				download: true,
				download_speed: this.speed
			}).then(streamInfo => {
				this.streamID = streamInfo.StreamID;
				this.$refs["playbackDownloadDlg"].download(start + " - " + end, streamInfo.StreamID);
			}).always(() => {
				this.bDownloading = false;
			});
		},
		toListView() {
			var rq = {};
			if(this.center) {
				rq.center = this.center;
			}
			if (this.indistinct) {
				rq.indistinct = this.indistinct;
			}
			this.$router.replace({
				path: `/devices/playback/list/${this.devid}/${this.channel}/${this.day}`,
				query: Object.assign({}, this.$route.query, rq),
			});
		},
		onTimeChange(video) {
			this.video = video;
		},
		onMediaInfo(mi) {
			this.mediaInfo = mi;
		},
		onEnded(e) {
			this.mediaInfo = null;
		},
		async onVideoError(e) {
			if(e == 'MediaError' && ((this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) || this.protocol == "HLS")) {
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
		onVideoTimeUpdate(t) {
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
			if(d < 3 && !this.switchTimer && endTime.format("YYYYMMDD") == this.day) {
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
		onPause() {
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
		onPlay() {
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
			if(this.streamID) {
				$.post("/api/v1/playback/control", {
					streamid: this.streamID,
					command: "scale",
					scale: speed,
				}).then(data => {
					this.speed = speed;
					this.speedTime = this.currentTime;
					this.speedAt = this.$refs["livePlayer"].getCurrentTime();
					this.$message({
                        type: "success",
                        message: `${speed} 倍速设置成功`,
					});
				})
			}
		},
		scaleOnStart() {
			if(this.streamID) {
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
			}
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
					this.downloadPlayback(t1.format("YYYY-MM-DDTHH:mm:ss"), t2.format("YYYY-MM-DDTHH:mm:ss"));
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
	}, //-- methods
	mounted() {
		// let mmt = moment();
		// let n = mmt.diff(mmt.clone().startOf('day'), 'minutes');
		// n -= 10;
		// if(n < 0) n = 0;
		// this.$refs.timeRule.clickMinute(n);
		// console.log(this.devid, this.channel, this.day)
		// this.getRecords(true);
		// $(window).on("beforeunload", this.beforeUnload);
		$(document).on("click", "#input-speed-menu a", this.onSpeedClick);
	},
	beforeDestroy() {
		$(window).off("beforeunload", this.beforeUnload);
		$(document).off("click", "#input-speed-menu a", this.onSpeedClick);
		this.stopPlayback();
	},
	beforeRouteLeave(to, from, next) {
		this.stopPlayback();
		next();
	},
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.center = to.query.center || "";
			vm.indistinct = to.query.indistinct || "";
			vm.ready();
		})
	},
	beforeRouteUpdate(to, from, next) {
		if (!to.params.day) {
			var rq = {};
			if(this.center) {
				rq.center = this.center;
			}
			if(this.indistinct) {
				rq.indistinct = this.indistinct;
			}
			next({
				path: `/devices/playback/${this.mode}/${to.params.devid}/${to.params.channel}/${moment().format('YYYYMMDD')}`,
				query: Object.assign({}, this.$route.query, rq),
				replace: true
			})
			return;
		} else {
            this.timerange = [
                moment(to.params.day, "YYYYMMDD").startOf('hour').toDate(),
                moment(to.params.day, "YYYYMMDD").startOf('hour').toDate()
            ]
        }
		this.center = to.query.center || "";
		this.indistinct = to.query.indistinct || "";
		this.stopPlayback();
		next();
		this.$nextTick(() => {
			if(!this.loading && from.path != to.path) {
				console.log(this.devid, this.channel, this.day)
				this.getRecords(true);
			}
		})
	}
};
</script>

<style lang="less" scoped>
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
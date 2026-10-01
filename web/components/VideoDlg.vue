<template>
<div :class="['modal', { fade: fade }]" data-keyboard="true" data-backdrop="static" tabindex="-1">
    <div class="modal-dialog modal-lgg">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                    <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title text-primary text-center"><span>{{videoTitle}}</span></h4>
            </div>
            <div class="modal-body">
                <div class="row">
                    <div :class="['play-area', {'user-active': active, 'col-sm-9': hasAnyRole(serverInfo, userInfo, '管理员', '操作员'), 'col-sm-12': !hasAnyRole(serverInfo, userInfo, '管理员', '操作员')}]" @mousemove="doActive">
                        <LivePlayer ref="player" v-if="bShow" :videoUrl="videoUrl" :waterMark="osd" :digitalZoom="digitalZoom"
                            :poster="snapUrl" :hideBigPlayButton="!!serverInfo.HideBigPlayButton" live :muted="!bAudioEnable"
                            @media_info="onMediaInfo" @ended="onEnded" @error="onError" @message="$message"
                            v-loading="bLoading" :loading.sync="bLoading" element-loading-text="加载中..." element-loading-background="#000">
                        </LivePlayer>
                    </div>
                    <div class="ptz-area col-sm-3" v-if="hasAnyRole(serverInfo, userInfo, '管理员', '操作员')">
                        <div class="form-group hide">
                            <div class="input-group">
                                <input type="text" v-model.trim="osd" class="form-control">
                                <span class="input-group-btn">
                                    <button class="btn btn-default" type="button" @click.prevent="setOSD">设置OSD</button>
                                </span>
                            </div>
                        </div>
                        <div class="form-group" style="margin:0px;">
                            <div class="ptz-block">
                                <el-progress class="mic-level-bar" :percentage="level" :text-inside="true" v-if="ws"></el-progress>
                                <div class="ptz-cell ptz-up" command="up" title="上" @mousedown.prevent="ptzControl('up', $event)" @touchstart.prevent="ptzControl('up', $event)">
                                    <i class="fa fa-chevron-up"></i>
                                </div>
                                <div class="ptz-cell ptz-left" command="left" title="左" @mousedown.prevent="ptzControl('left', $event)" @touchstart.prevent="ptzControl('left', $event)">
                                    <i class="fa fa-chevron-left"></i>
                                </div>
								<el-tooltip content="长按对讲" placement="top" manual :value="bTalkTip">
									<div class="ptz-cell ptz-center" title="云台控制">
										<i class="fa fa-microphone" title="按住喊话" @mousedown.prevent="talkStart"  @touchstart.prevent="talkStart" v-if="canTalk() && serverInfo.VersionType !== '标准版'"></i>
										<i class="fa fa-microphone-slash" title="由于浏览器安全策略, 非 HTTPS 或 localhost 访问, 对讲不可用" v-if="!canTalk() && serverInfo.VersionType !== '标准版'"></i>
									</div>
								</el-tooltip>
                                <div class="ptz-cell ptz-right" command="right" title="右" @mousedown.prevent="ptzControl('right', $event)" @touchstart.prevent="ptzControl('right', $event)">
                                    <i class="fa fa-chevron-right"></i>
                                </div>
                                <div class="ptz-cell ptz-down" command="down" title="下" @mousedown.prevent="ptzControl('down', $event)" @touchstart.prevent="ptzControl('down', $event)">
                                    <i class="fa fa-chevron-down"></i>
                                </div>
                                <div class="ptz-cell ptz-zoomin" command="zoomin" title="放大" @mousedown.prevent="ptzControl('zoomin', $event)" @touchstart.prevent="ptzControl('zoomin', $event)">
                                    <i class="fa fa-plus"></i>
                                </div>
                                <div class="ptz-cell ptz-zoomout" command="zoomout" title="缩小" @mousedown.prevent="ptzControl('zoomout', $event)" @touchstart.prevent="ptzControl('zoomout', $event)">
                                    <i class="fa fa-minus"></i>
                                </div>
                            </div>
                            <div class="clearfix"></div>
                            <el-tabs class="tab-info hidden-xs hidden-sm" v-model="activeName" v-if="SourceVideoCodecName && SourceVideoWidth && SourceVideoHeight">
                                <el-tab-pane  size="mini" label="流信息" name="steam-info">
                                    <div class="row stream-info hidden-xs hidden-sm">
                                        <div class="col-sm-3">
                                            分辨率:
                                        </div>
                                        <div class="col-sm-3">
                                            {{SourceVideoWidth}}x{{SourceVideoHeight}}
                                        </div>
                                        <div class="col-sm-3">
                                            帧率:
                                        </div>
                                        <div class="col-sm-3">
                                            {{FrameRate}}fps
                                        </div>
                                        <div class="clearfix">
											<div class="col-sm-3">
												码率:
											</div>
											<div class="col-sm-3">
												{{InBitRate}}Kbps
											</div>
											<div class="col-sm-3">
												丢包率:
											</div>
											<div class="col-sm-3">
												<a role="button" class="text-orange" @click="bRTPLostTip = !bRTPLostTip" v-if="RTPLostRate > 0 && Transport == 'UDP' && !isDemoUser(serverInfo, userInfo) && hasAnyRole(serverInfo, userInfo, '管理员')">
													<u>{{parseFloat(RTPLostRate).toFixed(2)}}% <i class="glyphicon glyphicon-exclamation-sign"></i></u>
												</a>
												<span :class="{'text-orange': RTPLostRate > 0}" v-else>
													{{parseFloat(RTPLostRate).toFixed(2)}}%
												</span>
											</div>
										</div>
                                        <div class="clearfix">
											<div class="col-sm-3">
												视频编码:
											</div>
											<div class="col-sm-3">
												{{SourceVideoCodecName}}
											</div>
											<div class="col-sm-3">
												音频编码:
											</div>
											<div :class="['col-sm-3', {'text-gray': !bAudioEnable}]">
												{{SourceAudioCodecName}}
											</div>
										</div>
                                        <div class="clearfix">
											<div class="col-sm-3">
												传输模式:
											</div>
											<div class="col-sm-3">
												{{Transport}}
											</div>
											<div class="col-sm-3">
												持续时长:
											</div>
											<div class="col-sm-3">
												{{formatDuration(Duration)}}
											</div>
										</div>
                                        <div class="clearfix" v-if="!isDemoUser(serverInfo, userInfo) && hasAnyRole(serverInfo, userInfo, '超级管理员')">
											<div class="col-sm-3">
												流地址:
											</div>
											<div class="col-sm-9">
												<i title="点击拷贝" role="button" v-clipboard="videoUrl" @success="$message({type:'success', message:'成功拷贝到粘贴板'})" class="fa fa-copy"></i>
												{{videoUrl}}
											</div>
										</div>
                                        <div class="clearfix" v-if="isDemoUser(serverInfo, userInfo)">
											<div class="col-sm-3 text-bold text-red">
												提示:
											</div>
											<div class="col-sm-9">
												演示系统限制匿名登录播放时间, 若需测试长时间播放, 请<a target="_blank" href="//www.liveqing.com/docs/download/LiveGBS.html">下载使用</a>
											</div>
										</div>
                                    </div>
                                </el-tab-pane>
                                <el-tab-pane  size="mini" label="预置位" name="preset-info">
                                    <Preset ref="presetList" :serial="serial" :code="code" :activeName="activeName"></Preset>
                                </el-tab-pane>
                            </el-tabs>
                        </div>
                    </div>
                </div>
            </div>
            <div class="modal-footer">
                <el-radio-group v-model.trim="protocol" size="small" @change="setProtocol" id="protocol-switcher" class="hidden-xs pull-left">
                    <el-radio-button label="FLV" v-if="flvSupported() && !isIE()">FLV</el-radio-button>
                    <el-radio-button label="WS_FLV" v-if="flvSupported()">WS_FLV</el-radio-button>
                    <el-radio-button label="WHEP" v-if="rtcSupported()">WEBRTC</el-radio-button>
                    <el-radio-button label="HLS">HLS</el-radio-button>
					<el-radio-button label="RTMP">RTMP</el-radio-button>
                </el-radio-group>
                <span class="text-red small" v-if="bOutHevcTip">
                    正在播放 H265 直出流, 确保浏览器版本较新, 并开启硬件加速&nbsp;&nbsp;
                </span>
                <span class="text-orange small hidden-xs hidden-sm hidden-md" v-else-if="bRTPLostTip && Transport == 'UDP'">
                    存在 UDP 丢包, 尝试返回设备列表将「流传输模式」设为「TCP 被动」并停止重新播放&nbsp;&nbsp;
                </span>
                <button v-if="streamid && hasAnyRole(serverInfo, userInfo, '管理员', '操作员') && serverInfo.VersionType !== '标准版'" type="button" :class="['btn', {'btn-primary': !bRecording, 'btn-danger': bRecording}]" @click.prevent="toggleRecord">
                    <i :class="['fa', {'fa-save': !bRecording, 'fa-stop': bRecording}]"></i>
                    {{bRecording? '停止录像' : '实时录像'}}
                </button>
                <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import "jquery-ui/ui/widgets/draggable";
import LivePlayer from "@liveqing/liveplayer";
import Preset from "components/Preset";

export default {
	data() {
		return {
			videoUrl: "",
			videoTitle: "",
			snapUrl: "",
			serial: "",
			code: "",
			timer: 0,
			active: false, // show demo tips
			osd: "",
			streamid: "",
			protocol: "",
			streamInfo: null,
			bRecording: false,
			bAudioEnable: false,
			recordFileURL: "",
			bShow: false,
			bLoading: false,
			recorder: null,
			ws: null,
			level: 0,
			statsTicker: 0,
			SourceVideoWidth: 0,
			SourceVideoHeight: 0,
			SourceVideoCodecName: "",
			SourceAudioCodecName: "",
			StartAt: "",
			Transport: "",
			InBitRate: 0,
			RTPLostRate: 0,
			FrameRate: 0,
			Duration: 0,
			NumOutputs: 0,
			activeName:"steam-info",
			mediaInfo: null,
			bOutHevcTip: false,
			bRTPLostTip: false,
			bTalkTip: false,
		};
	},
	props: {
		fade: {
			type: Boolean,
			default: false
		},
		digitalZoom: {
			type: Boolean,
			default: true
		},
		serverInfo: {
			type: Object,
			default: () => {}
		},
		userInfo: {
			type: Object,
			default: () => null
		}
	},
	beforeDestroy() {
		if(this.timer) {
			clearTimeout(this.timer);
			this.timer = 0;
		}
		if(this.statsTicker) {
			clearInterval(this.statsTicker);
			this.statsTicker = 0;
		}
		if(this.ws) {
			this.ws.close();
			this.ws = null;
		}
		this.ctrlStop();
		$(document).off("mouseup touchend", this.mouseUp);
	},
	mounted() {
		$(this.$el).find(".modal-content").draggable({
            handle: ".modal-header",
            cancel: ".modal-title span",
            addClasses: false,
            containment: "document",
            delay: 100,
            opacity: 0.5
        });
		$(this.$el).on("hide.bs.modal", () => {
			this.streamid = "";
			this.streamInfo = null;
			this.mediaInfo = null;
			this.bOutHevcTip = false;
			this.bRTPLostTip = false;
			this.bTalkTip = false;
			this.protocol = "";
			this.snapUrl = "";
			this.videoUrl = "";
			this.osd = "";
			this.bRecording = false;
			this.bAudioEnable = false;
			this.recordFileURL = "";
			this.activeName = "steam-info";
			if(this.timer) {
				clearTimeout(this.timer);
				this.timer = 0;
			}
			if(this.statsTicker) {
				clearInterval(this.statsTicker);
				this.statsTicker = 0;
			}
			this.ctrlStop();
			this.bShow = false;
		}).on("show.bs.modal", () => {
            this.bShow = true;
        });
		$(document).on("mouseup touchend", this.mouseUp);
	},
	components: { LivePlayer, Preset },
	methods: {
		play(title, serial, code, streamInfo) {
			streamInfo = streamInfo || {};
			var videoUrl = streamInfo.HLS;
			var protocol = "HLS";
			if(this.flvSupported()) {
                videoUrl = this.isIE() ? streamInfo.WS_FLV : streamInfo.FLV;
                protocol = this.isIE() ? "WS_FLV" : "FLV";
			}
			switch(String(this.serverInfo.PreferStreamFmt||'').toUpperCase()) {
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
			this.protocol = protocol;
			this.serial = serial || "";
			this.code = code || "";
			this.streamid = streamInfo.StreamID || "";
			this.streamInfo = streamInfo;
			this.readStreamInfo(streamInfo);
			this.videoTitle = title || "";
			this.osd = streamInfo.ChannelOSD || "";
			this.snapUrl = protocol == "RTMP" ? "" : (streamInfo.SnapURL || "");
			this.videoUrl = videoUrl || ""; // no need in next tick since player@2.6.9
			$(this.$el).modal("show");
			this.doActive();
			if(this.hasAnyRole(this.serverInfo, this.userInfo, '管理员', '操作员') && this.streamid) {
				this.statsTicker = setInterval(() => {
					this.updateStreamInfo();
				}, 3000);
			}
		},
		updateStreamInfo() {
			if(!this.streamid) return;
			$.ajax({
				method: "GET",
				url: "/api/v1/stream/info",
				global: false,
				data: {
					streamid: this.streamid,
				}
			}).then(streamInfo => {
				// proc stats data here
				// console.log(streamInfo);
				if(streamInfo) {
					this.streamInfo = streamInfo;
					this.readStreamInfo(streamInfo);
				}
			}).fail(xhr => {
				xhr && console.log(`update stream info ajax error: ${xhr.status} ${xhr.responseText}`);
			})
		},
		readStreamInfo(data) {
			data = data || {};
			if(!this.bRecording && !this.recordFileURL && data.RecordStartAt) {
				this.queryRecord();
			}
			this.bRecording = data.RecordStartAt != "";
			this.bAudioEnable = !!data.AudioEnable;
			this.SourceVideoWidth = data.SourceVideoWidth || 0;
			this.SourceVideoHeight = data.SourceVideoHeight || 0;
			this.SourceVideoCodecName = data.SourceVideoCodecName || "";
			if(this.SourceVideoCodecName == "hevc") {
				this.SourceVideoCodecName = "h265";
			}
			this.SourceAudioCodecName = data.SourceAudioCodecName || "";
			this.StartAt = data.StartAt || "";
			this.Transport = data.Transport || "";
			this.InBitRate = data.InBitRate || 0;
			this.FrameRate = data.SourceVideoFrameRate || 0;
			this.RTPLostRate = data.RTPLostRate || 0;
			this.Duration = data.Duration || 0;
			this.NumOutputs = data.NumOutputs || 0;
		},
		setProtocol(protocol) {
			if(!this.streamInfo) return;
			if(this.bOutHevcTip && !String(protocol).endsWith("FLV") && protocol != "HLS") {
				this.bOutHevcTip = false;
			}
			switch(protocol) {
				case "WEBRTC":
					this.videoUrl = this.streamInfo.WEBRTC;
					break;
				case "WHEP":
					this.videoUrl = this.streamInfo.WHEP;
					break;
				case "FLV":
					this.videoUrl = this.streamInfo.FLV;
					break;
				case "WS_FLV":
					this.videoUrl = this.streamInfo.WS_FLV;
					break;
				case "RTMP":
					this.snapUrl = "";
					this.videoUrl = this.streamInfo.RTMP;
					break;
				case "HLS":
					this.videoUrl = this.streamInfo.HLS;
					break;
			}
			this.doActive();
		},
		ptzControl(cmd, event) {
			// $.post("/api/v1/control/ptz", {
			//   serial: this.serial,
			//   code: this.code,
			//   command: cmd,
			// });
			$.ajax({
				method: "POST",
				url: "/api/v1/control/ptz",
				global: false,
				data: {
					serial: this.serial,
					code: this.code,
					command: cmd,
				}
			}).fail(xhr => {
				if(xhr) {
					console.log(`ptz ${cmd} ajax error: ${xhr.status} ${xhr.responseText}`);
					if(xhr.status == 401) {
						this.$nextTick(() => {
							location.href = `/login?r=${encodeURIComponent(window.location.href)}`;
						})
					}
				}
			})
			$(event.target).closest('.ptz-cell').addClass("active");
		},
		ptzStop() {
			if($(this.$el).find(".ptz-cell.active").length > 0) {
				// $.post("/api/v1/control/ptz", {
				//   serial: this.serial,
				//   code: this.code,
				//   command: "stop"
				// });
				$.ajax({
					method: "POST",
					url: "/api/v1/control/ptz",
					global: false,
					data: {
						serial: this.serial,
						code: this.code,
						command: "stop",
					}
				}).fail(xhr => {
					xhr && console.log(`ptz stop ajax error: ${xhr.status} ${xhr.responseText}`);
				})
				$(this.$el).find(".ptz-cell.active").removeClass("active");
			}
		},
		talkStart(e) {
			if(!this.ws) {
				this.ws = new WebSocket(this.wsTalkURL());
				this.ws.onopen = evt => {
					console.log("ws talk open");
				}
				this.ws.onclose = evt => {
					console.log("ws talk close");
				}
				this.ws.onerror = evt => {
					console.log("ws talk error", evt);
				}
			} else if(this.serverInfo.TalkHold) {
				this.bTalkTip = false;
				this.talkStop();
				return;
			}
			var $target = $(e.currentTarget);
			if(this.recorder) {
				this.bTalkTip = false;
				$target.addClass("active");
				this.recorder.start();
				return;
			}
			LiveRecorder.get((rec, err) => {
				if(err) {
					alert(err);
					this.talkStop();
					return
				}
				// this.$refs["player"].setMuted(true);
				this.bTalkTip = (!this.ws && !this.serverInfo.TalkHold);
				$target.addClass("active");
				this.recorder = rec;
				this.recorder.start();
			}, {
				sampleBits: 16,
				sampleRate: 8000,
				pcmCallback: (pcm, level) => {
					// binary to base64 string
					var reader = new window.FileReader();
					reader.onloadend = () => {
						var base64 = reader.result;
						var base64 = base64.split(',')[1];
						if(this.ws && this.ws.readyState === WebSocket.OPEN) {
							this.ws.send(base64);
							this.level = level;
						}
					}
					reader.readAsDataURL(pcm);
				}
			})
		},
		talkStop() {
			if(this.recorder) {
				// this.recorder.stop();
				this.recorder.destroy();
				this.recorder = null;
				this.bTalkTip = false;
				$(this.$el).find(".fa-microphone.active").removeClass("active");
				// this.$refs["player"].setMuted(false);
			}
			if(this.ws) {
				this.ws.close();
				this.ws = null;
			}
			this.level = 0;
		},
		ctrlStop() {
			this.talkStop();
			this.ptzStop();
		},
		mouseUp() {
			if(!this.serverInfo.TalkHold) {
				this.talkStop();
			}
			this.ptzStop();
		},
		toggleRecord() {
			if(this.bRecording) {
				$.ajax({
					method: "POST",
					url: "/api/v1/record/stop",
					global: false,
					data: {
						streamid: this.streamid
					}
				}).then(ret => {
					var fileURL = "";
					for(var record of ret.RecordList) {
						if(record.DownloadURL) {
							fileURL = record.DownloadURL;
							break;
						}
					}
					if(fileURL) {
						window.open(fileURL, "_blank");
					}
					this.bRecording = false;
					this.recordFileURL = "";
				}).fail(xhr => {
					xhr && console.log(`record stop ajax error: ${xhr.status} ${xhr.responseText}`);
					if(this.recordFileURL) {
						window.open(this.recordFileURL, "_blank");
						this.bRecording = false;
						this.recordFileURL = "";
					} else if(xhr && xhr.status) {
						let msg = xhr.responseText || "网络请求失败";
						try {
							msg = JSON.parse(msg)
						} catch (error) {}
						this.$message({
							type: 'error',
							message: msg
						})
					}
				})
			} else {
				$.post("/api/v1/record/start", {
					streamid: this.streamid
				}).then(ret => {
					this.bRecording = true;
					this.recordFileURL = ret.DownloadURL || "";
				})
			}
		},
		queryRecord() {
			$.ajax({
				method: "GET",
				url: "/api/v1/record/list",
				global: false,
				data: {
					streamid: this.streamid,
				}
			}).then(ret => {
				for(var record of ret.RecordList) {
					if(record.DownloadURL) {
						this.recordFileURL = record.DownloadURL;
						break;
					}
				}
			}).fail(xhr => {
				xhr && console.log(`record query ajax error: ${xhr.status} ${xhr.responseText}`);
			})
		},
		wsTalkURL() {
			var protocol = "ws:";
			if(location.protocol.startsWith("https")) {
				protocol = "wss:";
			}
			//url query param "format=pcm|g711a"
			return `${protocol}//${location.host}/api/v1/control/ws-talk/${this.serial}/${this.code}?format=pcm`;
		},
		doActive() {
            this.active = true;
            if(this.timer) {
                clearTimeout(this.timer);
                this.timer = 0;
            }
            this.timer = setTimeout(() => {
                this.active = false;
            }, 2000);
		},
		onMediaInfo(mi) {
			this.mediaInfo = mi;
		},
		onEnded(e) {
			this.mediaInfo = null;
		},
		onError(e) {
			if(e == 'MediaError' && ((this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) || this.protocol == "HLS")) {
				if(flvjs.getFeatureList() && !flvjs.getFeatureList().nativeMP4H265Playback) {
					this.bOutHevcTip = true;
					console.log("提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速");
				}
			}
		},
	}, //-- methods
};
</script>

<style lang="less" scoped>
.modal-title {
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;
}

.modal-body {
	padding: 10px;
}

.ptz-area {
	padding-left: 0;
}

.ptz-block {
	width: 150px;
	height: 200px;
	margin: 0 auto;
	text-align: center;
	position: relative;
	font-size: 24px;
	color: #333;

	.ptz-cell {
		width: 50px;
		height: 50px;
		line-height: 50px;
		position: absolute;
		cursor: pointer;
	}

	.ptz-cell.active {
		color: #ccc;
		// font-size: 26px;
	}

	.fa-microphone, .fa-microphone-slash {
		padding: 10px 13px;
		border-radius: 25px;
	}

	.fa-microphone.active {
		color: #FFF;
	}

	.mic-level-bar {
		position: absolute;
		transform: rotate(-90deg);
		width: 100%;
		top: 70px;
		left: -80px;
	}

	.ptz-center {
		top: 50px;
		left: 50px;
		border-radius: 25px;
		background-color: #ccc;
	}

	.ptz-up {
		top: 0;
		left: 50px;
	}

	.ptz-left {
		top: 50px;
		left: 0;
	}

	.ptz-right {
		top: 50px;
		left: 100px;
	}

	.ptz-down {
		top: 100px;
		left: 50px;
	}

	.ptz-zoomin {
		top: 150px;
		left: 20px;
	}

	.ptz-zoomout {
		top: 150px;
		left: 80px;
	}
}

.stream-info {
	line-height:24px;
	margin: 0px 0 0 0;
	font-size: 12px;
	.col-sm-3 {
		padding: 1px 0;
	}
	.col-sm-9 {
		padding: 1px;
		max-height: 120px;
		word-break: break-all;
		// &:hover {
		//   overflow-y: auto;
		// }
	}
}

#protocol-switcher {
	label {
		margin-bottom: 0;
	}
}

@media screen and(min-width: 768px) {
    .modal-dialog.modal-lgg {
        width: 90%;
    }
    .tab-info {
        display: none;
    }
}

@media screen and(min-width: 1200px) {
    .modal-dialog.modal-lgg {
        width: 1200px;
    }
    .tab-info {
        display: block;
    }
}
</style>

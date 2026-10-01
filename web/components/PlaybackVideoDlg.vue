<template>
<div :class="['modal', { fade: fade }]" data-keyboard="true" data-backdrop="static" tabindex="-1">
    <div class="modal-dialog modal-lg">
        <div class="modal-content">
            <div class="modal-header">
                <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                    <span aria-hidden="true">&times;</span>
                </button>
                <h4 class="modal-title text-primary text-center"><span>{{videoTitle}}</span></h4>
            </div>
            <div class="modal-body">
                <LivePlayer ref="livePlayer" v-if="bShow" :videoUrl="videoUrl" :muted="!bAudioEnable" :waterMark="osd" :digitalZoom="digitalZoom" hideLiveText live
                    :poster="snapUrl" :hideBigPlayButton="!!serverInfo.HideBigPlayButton"
                    @media_info="onMediaInfo" @ended="onEnded" @error="onError" @message="$message" @pause="onPause" @play="onPlay"
                    v-loading="bLoading" :loading.sync="bLoading" element-loading-text="加载中..." element-loading-background="#000">
                </LivePlayer>
                <div class="text-center" v-if="isDemoUser(serverInfo, userInfo) && !bOutHevcTip">
                    <br>
                    提示: 演示系统限制匿名登录播放时间, 若需测试长时间播放, 请<a target="_blank" href="//www.liveqing.com/docs/download/LiveGBS.html">下载使用</a>
                </div>
                <div class="text-center text-red" v-if="bOutHevcTip">
                    <br>
                    提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速
                </div>
            </div>
            <div class="modal-footer">
				<span v-show="inBitRate">{{inBitRate}}Kbps&nbsp;&nbsp;</span>
				<el-radio-group id="speed-switcher-playback" size="small" v-model.number="speed" @change="scale" v-show="streamID">
					<el-radio-button label="0.5">0.5x</el-radio-button>
					<el-radio-button label="1">1x</el-radio-button>
					<el-radio-button label="2">2x</el-radio-button>
					<el-radio-button label="4">4x</el-radio-button>
				</el-radio-group>
                <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import "jquery-ui/ui/widgets/draggable";
import LivePlayer from "@liveqing/liveplayer";

export default {
	data() {
		return {
			protocol: "",
			videoUrl: "",
			videoTitle: "",
			osd: "",
			snapUrl: "",
			serial: "",
			code: "",
			streamID: "",
			speed: 1,
			inBitRate: 0,
			timer: 0,
			bShow: false,
			bLoading: false,
			mediaInfo: null,
			bOutHevcTip: false,
			bAudioEnable: false,
			bPaused: false,
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
	components: { LivePlayer },
	mounted() {
		$(this.$el).find(".modal-content").draggable({
			handle: ".modal-header",
			cancel: ".modal-title span",
			addClasses: false,
			containment: "document",
			delay: 100,
			opacity: 0.5
		});
		$(this.$el).on("hidden.bs.modal", () => {
			this.bShow = false;
			if(this.timer) {
				clearInterval(this.timer);
				this.timer = 0;
			}
			this.stop();
			this.reset();
		}).on("shown.bs.modal", () => {
			this.bShow = true;
			if(this.streamID) {
				this.timer = setInterval(() => {
					$.ajax({
						type: "GET",
						url: "/api/v1/playback/streaminfo",
						data: {
                            streamid: this.streamID
						},
						global: false
					}).then(ret => {
						this.speed = ret.PlaybackSpeed || 1;
						this.inBitRate = ret.InBitRate || 0;
					}).fail(() => {
						this.inBitRate = 0;
					})
				}, 3000);
			}
		});
	},
	beforeDestroy() {
		if (this.timer) {
			clearInterval(this.timer);
			this.timer = 0;
		}
	},
	methods: {
		reset() {
			this.protocol = "";
			this.videoUrl = "";
			this.osd = "";
			this.snapUrl = "";
			this.serial = "";
			this.code = "";
			this.streamID = "";
			this.speed = 1;
			this.inBitRate = 0;
			this.timer = 0;
			this.mediaInfo = null;
			this.bOutHevcTip = false;
			this.bAudioEnable = false;
			this.bPaused = false;
		},
		play(protocol, src, title, snap, streamInfo) {
			this.protocol = protocol || "";
			this.videoTitle = title || "";
			this.snapUrl = snap || "";
			this.serial = streamInfo.DeviceID || "";
			this.code = streamInfo.ChannelID || "";
			this.streamID = streamInfo.StreamID || "";
			this.osd = streamInfo.ChannelOSD || "";
			this.speed = streamInfo.PlaybackSpeed || 1;
			this.inBitRate = streamInfo.InBitRate || 0;
			this.bAudioEnable = !!streamInfo.AudioEnable;
			this.videoUrl = src || ""; // no need in next tick since player@2.6.9
			$(this.$el).modal("show");
		},
		scale(speed = 1) {
			if(!this.streamID) return;
			$.post("/api/v1/playback/control", {
				streamid: this.streamID,
				command: "scale",
				scale: speed,
			}).then(data => {
				this.$message({
					type: "success",
					message: `${speed} 倍速设置成功`,
				});
			})
		},
		stop() {
			if(this.streamID) {
				$.ajax({
					type: "POST",
					url: "/api/v1/playback/stop",
					data: {
						streamid: this.streamID
					},
					global: false
				}).always(() => {
					this.$emit("close");
				})
				this.streamID = "";
			}
		},
		onMediaInfo(mi) {
			this.mediaInfo = mi;
		},
		onEnded(e) {
			this.mediaInfo = null;
			if(this.bShow && this.bPaused && this.streamID && this.videoUrl) {
				$.ajax({
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
				})
			}
		},
		onError(e) {
			if(e == 'MediaError' && ((this.mediaInfo && String(this.mediaInfo.videoCodec).startsWith("hvc")) || this.protocol == "HLS")) {
				if(flvjs.getFeatureList() && !flvjs.getFeatureList().nativeMP4H265Playback) {
					this.bOutHevcTip = true;
					console.log("提示: 正在播放 H265 直出流, 确保浏览器版本较新, 并且开启硬件加速");
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
						xhr && console.log(`pause stream[${this.streamID}] ajax error: ${xhr.status} ${xhr.responseText}`);
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
						xhr && console.log(`resume stream[${this.streamID}] ajax error: ${xhr.status} ${xhr.responseText}`);
					});
				}
			}
		},
	}
};
</script>

<style lang="less" scoped>
.modal-title {
	overflow: hidden;
	white-space: nowrap;
	text-overflow: ellipsis;
}

#speed-switcher-playback {
	margin-right: 10px;
	label {
		margin-bottom: 0;
	}
}
</style>

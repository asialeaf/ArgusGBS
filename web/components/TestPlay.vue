<template>
<div>
    <div class="input-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">协议模式
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-radio-group size="mini" v-model.trim="streamMode">
                    <el-radio label="UDP">UDP协议</el-radio>
                    <el-radio label="TCP active">TCP主动</el-radio>
                    <el-radio label="TCP passive">TCP被动</el-radio>
                </el-radio-group>
            </div>
        </div>
        <div class="form-group" v-show="!decode">
            <label class="col-sm-4 control-label">码流编号
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="streamnumber" v-model.trim="streamnumber">
                    <option value="">选择码流编号</option>
                    <option value="0">0-主码流</option>
                    <option value="1">1-子码流</option>
                </select>
            </div>
        </div>
        <div class="form-group" v-show="decode">
            <label for="input-code" class="col-sm-4 control-label">视频通道
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-select filterable clearable remote :remote-method="loadSrc" :loading="srcLoading" v-model="src"
                    @focus="!src && loadSrc();" ref="selectSrc"
                    placeholder="可搜索" size="small" style="width:100%;">
                    <el-option v-for="item in srcList" :key="`${item.DeviceID}:${item.ID}`" :label="item.ID" :value="`${item.DeviceID}:${item.ID}`">
                    </el-option>
                </el-select>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="start" :disabled="starting||!!streamInfo">
                    {{ decode ? '解码上屏' : '点播' }}
                    <span v-show="starting">...</span></button>
                <button role="button" class="btn btn-danger" @click.prevent="stop" :disabled="stopping" v-if="!!streamInfo">断开<span v-show="stopping">...</span></button>
            </div>
        </div>
    </div>
    <div class="play-area" v-if="videoUrl">
        <div class="row">
            <div class="col-sm-10 col-sm-offset-1">
                <LivePlayer ref="player" :videoUrl="videoUrl" live stretch :muted="!audio" :controls="false"></LivePlayer>
            </div>
        </div>
    </div>
    <div class="info-area" v-if="SourceVideoCodecName && SourceVideoWidth && SourceVideoHeight">
        <div class="row col-sm-10 col-sm-offset-1 no-padding">
            <div class="col-md-3">
                分辨率:
            </div>
            <div class="col-md-3">
                {{SourceVideoWidth}}x{{SourceVideoHeight}}
            </div>
            <div class="col-md-3">
                帧率:
            </div>
            <div class="col-md-3">
                {{FrameRate}}fps
            </div>
            <div class="clearfix">
                <div class="col-md-3">
                    码率:
                </div>
                <div class="col-md-3">
                    {{InBitRate}}Kbps
                </div>
                <div class="col-md-3">
                    丢包率:
                </div>
                <div class="col-md-3">
                    <span>
                        {{parseFloat(RTPLostRate).toFixed(2)}}%
                    </span>
                </div>
            </div>
            <div class="clearfix">
                <div class="col-md-3">
                    视频编码:
                </div>
                <div class="col-md-3">
                    {{SourceVideoCodecName}}
                </div>
                <div class="col-md-3">
                    音频编码:
                </div>
                <div class="col-md-3">
                    {{SourceAudioCodecName}}
                </div>
            </div>
            <div class="clearfix">
                <div class="col-md-3">
                    传输模式:
                </div>
                <div class="col-md-3">
                    {{Transport}}
                </div>
                <div class="col-md-3">
                    持续时长:
                </div>
                <div class="col-md-3">
                    {{formatDuration(Duration)}}
                </div>
            </div>
        </div>
    </div>
</div>
</template>

<script>
import LivePlayer from "@liveqing/liveplayer";

export default {
    data() {
        return {
            streamMode: "UDP",
            streamnumber: "",
            starting: false,
            stopping: false,
            audio: false,
            streamInfo: null,
            videoUrl: "",
            ticker: 0,
            src: "",
            srcList: () => [],
            srcLoading: false,

            SourceVideoCodecName: "",
            SourceAudioCodecName: "",
            SourceVideoWidth: 0,
            SourceVideoHeight: 0,
            FrameRate: 0,
            InBitRate: 0,
            RTPLostRate: 0,
            Transport: "",
            Duration: 0,
        }
    },
    computed: {
        decode() {
            var serial = this.$store.state.serial;
            if(serial && serial.length >= 13 && serial.substr(10, 3) == "114") {
                return true;
            }
            var code = this.$store.state.code;
            if(code && code.length >= 13 && code.substr(10, 3) == "133") {
                return true;
            }
            return false;
        },
    },
    watch: {
        decode: function(newVal, oldVal) {
            if(newVal && this.srcList.length == 0) {
                this.loadSrc();
            }
        }
    },
    components: { LivePlayer },
    methods: {
        async start() {
            this.starting = true;
            await this.$store.dispatch("connect");

            var transport = "UDP";
            var transport_mode = "passive";
            switch(this.streamMode) {
                case "TCP passive":
                    transport = "TCP";
                    transport_mode = "passive";
                    break;
                case "TCP active":
                    transport = "TCP";
                    transport_mode = "active";
                    break;
            }
            if(this.decode && !this.src) {
                this.$message.error("请选择视频通道");
                this.starting = false;
                return;
            }
            if(this.decode) {
                $.post("/api/v1/decode/start", {
                    serial: this.$store.state.serial,
                    code: this.$store.state.code,
                    transport: transport,
                    transport_mode: transport_mode,
                    streamid: `stream:${this.src}`,
                }).then((streamInfo = {}) => {
                    this.streamInfo = streamInfo;
                    this.readStreamInfo(streamInfo);
                }).always(() => {
                    this.starting = false;
                })
            } else {
                $.post("/api/v1/stream/start", {
                    serial: this.$store.state.serial,
                    code: this.$store.state.code,
                    transport: transport,
                    transport_mode: transport_mode,
                    streamnumber: this.streamnumber,
                }).then((streamInfo = {}) => {
                    this.streamInfo = streamInfo;
                    this.readStreamInfo(streamInfo);
                }).always(() => {
                    this.starting = false;
                })
            }
        },
        stop() {
            this.stopping = true;
            var target = "/api/v1/stream/stop";
            if(this.decode) {
                target = "/api/v1/decode/stop";
            }
			$.post(target, {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
			}).then(ret => {
                this.streamInfo = null;
                this.readStreamInfo(null);
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
			}).always(() => {
				this.stopping = false;
			})
        },
		readStreamInfo(data) {
			data = data || {};
			this.audio = !!data.AudioEnable;
			this.SourceVideoWidth = data.SourceVideoWidth || 0;
			this.SourceVideoHeight = data.SourceVideoHeight || 0;
			this.SourceVideoCodecName = data.SourceVideoCodecName || "";
			if(this.SourceVideoCodecName == "hevc") {
				this.SourceVideoCodecName = "h265";
			}
			this.SourceAudioCodecName = data.SourceAudioCodecName || "";
			this.Transport = data.Transport || "";
			this.InBitRate = data.InBitRate || 0;
			this.FrameRate = data.SourceVideoFrameRate || 0;
			this.RTPLostRate = data.RTPLostRate || 0;
			this.Duration = data.Duration || 0;

            var videoUrl = data.HLS || "";
            if(this.flvSupported()) {
                videoUrl = this.isIE() ? data.WS_FLV || "" : data.FLV || "";
            }
            this.videoUrl = videoUrl;
		},
        updateStreamInfo() {
            var serial = this.$store.state.serial;
            var code = this.$store.state.code;
            if(!serial||!code) return;
            if(this.decode) {
                $.ajax({
                    method: "GET",
                    url: "/api/v1/decode/info",
                    global: false,
                    data: {
                        serial: serial,
                        code: code,
                    }
                }).then(streamInfo => {
                    if(streamInfo) {
                        this.streamInfo = streamInfo;
                        this.readStreamInfo(streamInfo);
                    }
                })
            } else {
                $.ajax({
                    method: "GET",
                    url: "/api/v1/stream/info",
                    global: false,
                    data: {
                        streamid: `stream:${serial}:${code}`,
                    }
                }).then(streamInfo => {
                    if(streamInfo) {
                        this.streamInfo = streamInfo;
                        this.readStreamInfo(streamInfo);
                    }
                })
            }
        },
        loadSrc(q = "") {
            this.srcLoading = true;
            $.get("/api/v1/device/channellist", {
                q: q == this.src ? '' : q,
                channel_type: "device",
                online: true,
                limit: 10,
            }).then(ret => {
                this.srcList = ret.ChannelList || [];
                if(this.srcList.length > 0 && !this.src && !q) {
                    this.src = `${this.srcList[0].DeviceID}:${this.srcList[0].ID}`;
                }
            }).always(() => {
                this.srcLoading = false;
            })
        },
    },
    mounted() {
        if(this.decode) {
            this.loadSrc();
        }
        this.updateStreamInfo();
        this.ticker = setInterval(() => {
            this.updateStreamInfo();
        }, 3000);
    },
    beforeDestroy() {
        if(this.ticker) {
            clearInterval(this.ticker);
            this.ticker = 0;
        }
    },
}
</script>

<style lang="less" scoped>
.el-radio-group {
    margin-top: 9px;

    .el-radio + .el-radio {
        margin-left: 15px;
    }
}

.play-area {
    margin: 50px 0 10px 0;
}

.info-area {
    margin: 10px 0;
}
</style>

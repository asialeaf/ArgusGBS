<template>
<div>
    <div class="query-area form-horizontal" v-show="!videoUrl">
        <div class="form-group">
            <label class="col-sm-4 control-label">起始时间
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-date-picker type="datetime" size="small" :picker-options="pickerOptions" v-model="queryStartTime"></el-date-picker>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">结束时间
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <el-date-picker type="datetime" size="small" :picker-options="pickerOptions" v-model="queryEndTime"></el-date-picker>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">录像类型
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="queryRecordType" v-model="queryRecordType">
                    <option value="">不选</option>
                    <option value="time">定时录像</option>
                    <option value="alarm">报警录像</option>
                    <option value="manual">手动录像</option>
                    <option value="all">所有</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">码流编号
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="queryStreamNumber" v-model="queryStreamNumber">
                    <option value="">不选</option>
                    <option value="0">主码流</option>
                    <option value="1">子码流</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <div class="col-sm-offset-4 col-sm-7">
                <button role="button" class="btn btn-primary" @click.prevent="query" :disabled="loading">录像检索<span v-show="loading">...</span></button>
                <el-badge :value="records.length" :hidden="records.length <= 0">
                    <button role="button" class="btn btn-info" @click.prevent="showRecords">查看结果</button>
                </el-badge>
            </div>
        </div>
    </div>
    <div class="play-area" v-show="videoUrl">
        <div class="row">
            <div class="col-sm-10 col-sm-offset-1 col-lg-8 col-lg-offset-2">
                <LivePlayer ref="player" :videoUrl="videoUrl" live stretch :muted="!audio" :controls="false" :liveInterval="5000" hideBigPlayButton>
                </LivePlayer>
            </div>
        </div>
    </div>
    <ul class="nav nav-tabs" v-show="!streamInfo">
        <li class="active">
            <a href="#playback-area" data-toggle="tab">录像回放</a>
        </li>
        <li>
            <a href="#download-area" data-toggle="tab">下载</a>
        </li>
    </ul>
    <br>
    <div class="common-area form-horizontal">
        <div class="form-group">
            <label class="col-sm-4 control-label">开始时间
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="playbackStartTime" v-model="playbackStartTime" placeholder="选择录像自动获取">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">结束时间
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <input type="text" class="form-control" name="playbackEndTime" v-model="playbackEndTime" placeholder="选择录像自动获取">
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">协议模式
                <span class="text-red">*</span>
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="playbackTransportMode" v-model="playbackTransportMode">
                    <option value="UDP">UDP协议</option>
                    <option value="TCP active">TCP主动</option>
                    <option value="TCP passive">TCP被动</option>
                </select>
            </div>
        </div>
        <div class="form-group">
            <label class="col-sm-4 control-label">码流编号
            </label>
            <div class="col-sm-7">
                <select class="form-control" name="playbackStreamNumber" v-model="playbackStreamNumber">
                    <option value="">选择码流编号</option>
                    <option value="0">主码流</option>
                    <option value="1">子码流</option>
                </select>
            </div>
        </div>
    </div>
    <div class="tab-content">
        <div class="tab-pane active" id="playback-area">
            <div class="form-horizontal">
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <el-badge :value="records.length" :hidden="records.length <= 0">
                            <button role="button" class="btn btn-info" @click.prevent="selectRecords" :disabled="!!streamInfo">选择文件</button>
                        </el-badge>
                        &nbsp;&nbsp;
                        <button role="button" class="btn btn-primary" @click.prevent="playbackStart(false)" :disabled="starting||!!streamInfo">回放<span v-show="starting||!!streamInfo">...</span></button>
                        <button role="button" class="btn btn-danger" @click.prevent="playbackStop" :disabled="stopping" v-if="!!streamInfo">停止<span v-show="stopping">...</span></button>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-12 text-center">
                        <div class="btn-group">
                            <button type="button" class="btn btn-default btn-sm" :disabled="controling||!streamInfo" @click.prevent="playbackControl('scale', '', scale <= 1 ? scale/2 : 0.5)"><i class="fa fa-backward"></i> 慢放</button>
                            <button type="button" class="btn btn-default btn-sm" :disabled="controling||!streamInfo" @click.prevent="playbackControl('scale', '', scale >= 1 ? scale*2 : 2)"><i class="fa fa-forward"></i> 快放</button>
                            <button type="button" class="btn btn-default btn-sm" :disabled="controling||!streamInfo" @click.prevent="playbackControl('scale', '', 1)"><i class="glyphicon glyphicon-repeat"></i> 正常速</button>
                            <button type="button" class="btn btn-default btn-sm" :disabled="controling||!streamInfo" @click.prevent="playbackControl('play', '', scale)" v-if="paused"><i class="fa fa-play-circle-o"></i> 继续</button>
                            <button type="button" class="btn btn-default btn-sm" :disabled="controling||!streamInfo" @click.prevent="playbackControl('pause', '', scale)" v-else><i class="fa fa-pause-circle-o"></i> 暂停</button>
                            <button type="button" class="btn btn-default btn-sm" :disabled="controling||!streamInfo" @click.prevent="playbackControl('teardown', '', scale)"><i class="fa fa-times-circle-o"></i> 停止</button>
                        </div>
                    </div>
                </div>
                <div class="form-group">
                    <label class="col-sm-2 control-label">随机拖动</label>
                    <div class="col-sm-9">
                        <el-slider v-model.number="npt" :min="0" :max="playbackDuration" :format-tooltip="formatNPT" :disabled="controling||!streamInfo" @change="npt => playbackControl('play', npt, scale)"></el-slider>
                    </div>
                </div>
            </div>
        </div> <!-- playback-area -->
        <div class="tab-pane" id="download-area">
            <div class="form-horizontal">
                <div class="form-group">
                    <label class="col-sm-4 control-label">下载倍速
                        <span class="text-red">*</span>
                    </label>
                    <div class="col-sm-7">
                        <el-radio-group size="mini" v-model.number="download_speed" @change="scale => playbackControl('scale', '', scale)">
                            <el-radio :label="1">x1</el-radio>
                            <el-radio :label="2">x2</el-radio>
                            <el-radio :label="4">x4</el-radio>
                        </el-radio-group>
                    </div>
                </div>
                <div class="form-group">
                    <div class="col-sm-offset-4 col-sm-7">
                        <el-badge :value="records.length" :hidden="records.length <= 0">
                            <button role="button" class="btn btn-info" @click.prevent="selectRecords" :disabled="!!streamInfo">选择文件</button>
                        </el-badge>
                        &nbsp;&nbsp;
                        <button role="button" class="btn btn-primary" @click.prevent="playbackStart(true)" :disabled="downloading||!!streamInfo">下载<span v-show="downloading||!!streamInfo">...</span></button>
                        <button role="button" class="btn btn-danger" @click.prevent="playbackStop" :disabled="stopping" v-if="!!streamInfo">停止<span v-show="stopping">...</span></button>
                    </div>
                </div>
                <div class="form-group" v-show="fileUrl && streamInfo">
                    <label class="col-sm-4 control-label">下载进度</label>
                    <div class="col-sm-7">
                        <el-progress :text-inside="true" :stroke-width="18" :percentage="progress"></el-progress>
                        <p v-if="fileError" class="text-red text-center">
                            <br>
                            <span class="text-bold">异常终止提示：</span>
                            {{fileError}}
                        </p>
                        <p v-else-if="inBitRate">
                            <br>
                            <span>{{inBitRate}}Kbps</span>
                        </p>
                    </div>
                </div>
            </div>
        </div> <!-- download-area -->
    </div> <!-- tab-content -->
    <PlaybackListDlg ref="playbackListDlg" :records="records" @selected="onSelected"></PlaybackListDlg>
</div>
</template>

<script>
import moment from 'moment';
import LivePlayer from "@liveqing/liveplayer";
import PlaybackListDlg from 'components/PlaybackListDlg.vue';

export default {
    data() {
        return {
            pickerOptions: {
                shortcuts: [{
                    text: '半小时前',
                    onClick(picker) {
                        picker.$emit('pick', new Date(new Date().getTime() - 1800000));
                    },
                }, {
                    text: '一小时前',
                    onClick(picker) {
                        picker.$emit('pick', new Date(new Date().getTime() - 3600000));
                    },
                }, {
                    text: '六小时前',
                    onClick(picker) {
                        picker.$emit('pick', new Date(new Date().getTime() - 3600000 * 6));
                    },
                }, {
                    text: '一天前',
                    onClick(picker) {
                        picker.$emit('pick', new Date(new Date().getTime() - 86400000));
                    },
                }, {
                    text: '一周前',
                    onClick(picker) {
                        picker.$emit('pick', new Date(new Date().getTime() - 86400000 * 7));
                    },
                }]
            },
            queryStartTime: new Date(new Date().getTime() - 3600000),
            queryEndTime: new Date(),
            queryRecordType: "",
            queryStreamNumber: "",
            loading: false,
            records: [],

            playbackStartTime: "",
            playbackEndTime: "",
            playbackTransportMode: "UDP",
            playbackStreamNumber: "",
            playbackDuration: 100,
            starting: false,
            downloading: false,
            stopping: false,
            controling: false,
            paused: true,
            npt: 0,
            scale: 1,
            download_speed: 1,
            ticker: 0,

            videoUrl: "",
            fileUrl: "",
            fileError: "",
            progress: 0,
            inBitRate: 0,
            audio: false,
            streamInfo: null,
        }
    },
    components: { LivePlayer, PlaybackListDlg },
    methods: {
        async query() {
            this.loading = true
            await this.$store.dispatch('connect', {
                topic: "录像检索"
            });

            $.get("/api/v1/playback/recordlist", {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                starttime: moment(this.queryStartTime).format("YYYY-MM-DDTHH:mm:ss"),
                endtime: moment(this.queryEndTime).format("YYYY-MM-DDTHH:mm:ss"),
                type: this.queryRecordType,
                streamnumber: this.queryStreamNumber,
            }).then(ret => {
                this.records = ret.RecordList||[];
                this.$message({
                    type: "success",
                    message: `录像检索完成, 共检索到${this.records.length}条录像文件`
                })
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
            }).always(() => {
                this.loading = false;
            })
        },
        showRecords() {
            this.$refs['playbackListDlg'].show();
        },
        selectRecords() {
            this.$refs['playbackListDlg'].show(true);
        },
        onSelected(row) {
            if(!row) return;
            this.playbackStartTime = row.StartTime||"";
            this.playbackEndTime = row.EndTime||"";
        },
        async playbackStart(download = false) {
            if(!this.playbackStartTime || !this.playbackEndTime) {
                this.$message({
                    type: "error",
                    message: "请先选择录像文件"
                })
                return;
            }
            this.starting = true;
            await this.$store.dispatch('connect', {
                topic: download ? "录像下载" : "录像回放"
            });

            var transport = "UDP";
            var transport_mode = "passive";
            switch(this.playbackTransportMode) {
                case "TCP passive":
                    transport = "TCP";
                    transport_mode = "passive";
                    break;
                case "TCP active":
                    transport = "TCP";
                    transport_mode = "active";
                    break;
            }
            $.post("/api/v1/playback/start", {
				serial: this.$store.state.serial,
				code: this.$store.state.code,
                starttime: this.playbackStartTime,
                endtime: this.playbackEndTime,
                transport: transport,
                transport_mode: transport_mode,
                streamnumber: this.playbackStreamNumber,
                download: download,
                download_speed: this.download_speed,
            }).then(streamInfo => {
                this.audio = !!streamInfo.AudioEnable;
                this.streamInfo = streamInfo;
                var videoUrl = streamInfo.HLS || "";
                if(this.flvSupported()) {
                    videoUrl = this.isIE() ? streamInfo.WS_FLV || "" : streamInfo.FLV || "";
                }
                this.videoUrl = videoUrl;
                this.paused = false;
                this.playbackDuration = streamInfo.PlaybackDuration || 100;
                this.fileUrl = streamInfo.PlaybackFileURL || "";
                if(this.fileUrl) {
                    this.progress = Math.ceil(streamInfo.Progress * 100);
                    if(this.ticker) {
                        clearInterval(this.ticker);
                        this.ticker = 0;
                    }
                    this.ticker = setInterval(() => {
                        $.ajax({
                            type: "GET",
                            url: "/api/v1/playback/streaminfo",
                            data: {
                                streamid: streamInfo.StreamID,
                            },
                            global: false
                        }).then(ret => {
                            this.progress = Math.ceil(ret.Progress * 100);
                            this.fileUrl = ret.PlaybackFileURL || "";
                            this.fileError = ret.PlaybackFileError || "";
                            this.inBitRate = ret.InBitRate || 0;
                        }).fail(() => {
                            this.progress = 100;
                            this.inBitRate = 0;
                        })
                    }, 3000);
                }
            }).always(() => {
                this.starting = false;
            })
        },
        async playbackControl(command, range, scale) {
            var topic = "";
            switch(command) {
                case "play":
                    if(range === '' || range === "now") {
                        topic = "断续播放";
                    } else {
                        topic = "随机播放";
                    }
                    break;
                case "pause":
                    topic = "暂停播放";
                    this.$refs["player"].snap();
                    break;
                case "scale":
                    topic = "设置倍速";
                    break;
                case "teardown":
                    topic = "停止";
                    break;
            }
            if(!topic || !this.streamInfo) return;
            this.controling = true;
            await this.$store.dispatch('connect');

            $.post("/api/v1/playback/control", {
                streamid: this.streamInfo.StreamID,
                command: command,
                range: range,
                scale: scale
            }).then(ret => {
                this.$message({
                    type: "success",
                    message: `录像回放${topic}成功`
                })
                if(command == "pause") {
                    this.paused = true;
                    // this.$refs["player"].pause();
                } else if(command == "teardown") {
                    this.paused = true;
                    // setTimeout(() => {
                    //     this.playbackStop();
                    // }, 1000);
                } else {
                    if(command == "scale" && scale) {
                        this.scale = scale;
                    }
                    if(command == "play" && range !== "" && range !== "now") {
                        this.npt = range;
                    }
                    this.paused = false;
                    // this.$refs["player"].play();
                }
            }).always(() => {
                this.controling = false;
            })
        },
        playbackStop() {
            if(!this.streamInfo) return;
            this.stopping = true;
			$.post("/api/v1/playback/stop", {
				streamid: this.streamInfo.StreamID,
			}).then(ret => {
                this.videoUrl = "";
                this.streamInfo = null;
                this.audio = false;
                this.paused = true;
                this.npt = 0;
                this.playbackDuration = 100;
                this.inBitRate = 0;
                this.fileError = "";
                this.progress = 0;
                if(this.ticker) {
                    clearInterval(this.ticker);
                    this.ticker = 0;
                }
                if(this.fileUrl) {
                    window.open(this.fileUrl, "_blank");
                    this.fileUrl = "";
                }
                setTimeout(() => {
                    this.$store.dispatch("disconnect");
                }, 1000);
			}).always(() => {
				this.stopping = false;
			})
        },
        formatNPT(npt) {
            return `${npt} 秒`;
        },
        beforeUnload(event) {
            this.reset();
            event.returnValue = '';
        },
        reset() {
            if(this.ticker) {
                clearInterval(this.ticker);
                this.ticker = 0;
            }
            this.playbackStop();
        },
    },
    mounted() {
        $(window).on("beforeunload", this.beforeUnload);
    },
    beforeDestroy() {
        $(window).off("beforeunload", this.beforeUnload);
        this.reset();
    },
}
</script>

<style lang="less" scoped>
.el-date-editor {
    width: 100%;
}

.nav.nav-tabs > li {
    a {
        background-color: transparent;
    }
    &.active {
        a {
            background-color: darken(#ecf0f5, 5%);
        }
    }
}

.el-radio-group {
    margin-top: 9px;

    .el-radio + .el-radio {
        margin-left: 15px;
    }
}

.el-progress {
    margin-top: 10px;
}
</style>
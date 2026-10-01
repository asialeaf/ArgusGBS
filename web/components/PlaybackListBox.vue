<template>
	<div class="">
		<div class="box box-primary">
			<div class="box-header">
			    <h4 class="text-primary text-center">设备录像列表({{devid}}-{{channel}})</h4>
			</div>
			<div class="box-body">
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group form-group-sm">
                        <button type="button" class="btn btn-sm btn-primary" @click.prevent="$router.go(-1)" v-if="$getQueryString('menu') != 'no'">
                            <i class="fa fa-chevron-left"></i> 返回
                        </button>
                    </div>
                    <div class="form-group pull-right">
                        <div class="input-group input-group-sm">
                            <DatePicker class="form-control" @update:day="updateDay" :day="day" ref="datePicker"></DatePicker>
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-default" @click.prevent="showDatePicker">
                                    <i class="fa fa-calendar"></i>
                                </button>
                                <button @click.prevent="toTimeView" class="btn btn-default btn-sm">
                                    <i class="fa fa-hand-o-right"></i> 时间轴视图
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
                <br>
                <div class="clearfix"></div>
                <el-table stripe v-loading="loading" element-loading-text="加载中..."
                    ref="playbackTable" :data="pageData" :default-sort="{prop: 'StartTime', order: 'ascending'}" @sort-change="sortChange">
                    <el-table-column prop="DeviceID" label="通道国标编号" min-width="200" show-overflow-tooltip></el-table-column>
                    <el-table-column label="操作" min-width="190" :fixed="isMobile() ? false : 'right'">
                    <template slot-scope="props">
                        <div class="btn-group btn-group-xs">
                            <button type="button" class="btn btn-primary" @click.prevent="startPlayback(props.row)" :disabled="props.row.Starting"><i class="fa fa-play-circle"></i> 播放</button>
                            <button type="button" class="btn btn-info" @click.prevent="downloadPlayback(props.row)" :disabled="props.row.Starting"><i class="fa fa-download"> 下载</i></button>
                            <a :href="`/play.html?type=playback&ptz=no&serial=${devid}&code=${props.row.DeviceID}&starttime=${props.row.StartTime}&endtime=${props.row.EndTime}`" role="button" class="btn btn-warning" target="_blank"><i class="fa fa-share"></i> 分享页</a>
                        </div>
                    </template>
                    </el-table-column>
                    <el-table-column prop="Name" label="通道名称" min-width="120" :formatter="formatName" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="StartTime" label="开始时间" min-width="160" :formatter="formatName" sortable></el-table-column>
                    <el-table-column prop="EndTime" label="结束时间" min-width="160" :formatter="formatName" sortable></el-table-column>
                    <el-table-column prop="FileSize" label="文件大小" min-width="100" :formatter="formatFileSize" align="right" sortable></el-table-column>
                </el-table>
			</div>
			<div class="box-footer clearfix" v-if="total > 0">
			    <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
			</div>
		</div>
		<PlaybackVideoDlg ref="playbackVideoDlg" :serverInfo="serverInfo" :userInfo="userInfo" @close="onClose"></PlaybackVideoDlg>
		<PlaybackDownloadDlg ref="playbackDownloadDlg" @download="onDownload"></PlaybackDownloadDlg>
	</div>
</template>

<script>
import { mapState } from "vuex";
import _ from "lodash";
import moment from "moment";
import prettyBytes from 'pretty-bytes';
import DatePicker from "components/DatePicker.vue";
import PlaybackVideoDlg from "components/PlaybackVideoDlg.vue";
import PlaybackDownloadDlg from "components/PlaybackDownloadDlg.vue";

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
            default: "list"
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
            currentPage: 1,
            pageSize: 10,
            loading: false,
            sort: "",
            order: "",
            center: "",
            indistinct: "",
            streamID: "",
            records: []
        };
	},
	components: {
	    PlaybackVideoDlg, PlaybackDownloadDlg, DatePicker
	},
	computed: {
        ...mapState(['userInfo', 'serverInfo']),
        total() {
            return this.records.length;
        },
        pageData() {
            let start = (this.currentPage - 1) * this.pageSize;
            let end = start + this.pageSize;
            return this.records.slice(start, end);
        }
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
                this.currentPage = 1;
            }
            if(!this.nextTimeRange()){
                this.sortRecords();
                this.$nextTick(() => {
                    this.loading = false;
                })
                return
            }
            $.ajax("/api/v1/playback/recordlist", {
                type: 'get',
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
        formatName(row, col, cell) {
            return cell || "-";
        },
        formatFileSize(row, col, cell) {
            return cell ? prettyBytes(cell) : "-";
        },
        stopPlayback() {
            return new Promise((resolve, reject) => {
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
            })
        },
        beforeUnload(event) {
            this.stopPlayback();
            // event.preventDefault();
            event.returnValue = '';
        },
        downloadPlayback(row) {
            this.loading = true;
            this.$set(row, "Starting", true);
            $.post("/api/v1/playback/start", {
                serial: this.devid,
                code: this.channel,
                starttime: row.StartTime,
                endtime: row.EndTime,
                download: true
            }).then(streamInfo => {
                this.streamID = streamInfo.StreamID;
                this.$refs["playbackDownloadDlg"].download(row.StartTime + " - " + row.EndTime, streamInfo.StreamID);
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
                this.$delete(row, "Starting");
            });
        },
        startPlayback(row) {
            this.loading = true;
            this.$set(row, "Starting", true);
            $.post("/api/v1/playback/start", {
                serial: this.devid,
                code: this.channel,
                starttime: row.StartTime,
                endtime: row.EndTime
            }).then(streamInfo => {
                this.streamID = streamInfo.StreamID;
                var videoUrl = streamInfo.HLS;
                var protocol = "HLS";
                if(this.flvSupported()) {
                    protocol = this.isIE() ? "WS_FLV" : "FLV";
                    videoUrl = this.isIE() ? streamInfo.WS_FLV : streamInfo.FLV;
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
                var snap = protocol == "RTMP" ? "" : (row.Snap || "");
                this.$refs["playbackVideoDlg"].play(
                    protocol,
                    videoUrl,
                    row.Name || row.DeviceID,
                    snap,
                    streamInfo
                );
            }).always(() => {
                this.$nextTick(() => {
                    this.loading = false;
                })
                this.$delete(row, "Starting");
            });
        },
        toTimeView() {
            var rq = {};
            if(this.center) {
                rq.center = this.center;
            }
            if(this.indistinct) {
                rq.indistinct = this.indistinct;
            }
            this.$router.replace({
                path: `/devices/playback/timebox/${this.devid}/${this.channel}/${this.day}`,
                query: Object.assign({}, this.$route.query, rq),
            });
        },
        sortChange(data) {
            if(!data || !data.prop || !data.order) return;
            this.sort = data.prop;
            this.order = data.order;
            this.sortRecords();
        },
        sortRecords() {
            if(!this.sort) return;
            this.records.sort((x, y) => {
                var t1 = x[this.sort];
                var t2 = y[this.sort];
                var ret = 0;
                if(!t1 || !t2) return ret;
                if(t1 < t2) {
                    ret = -1;
                } else if(t1 > t2) {
                    ret = 1;
                }
                if(this.order === "desc" || this.order === "descending") {
                    ret = -1 * ret;
                }
                return ret;
            })
        },
        onDownload() {
            this.streamID = "";
        },
        onClose() {
            this.streamID = "";
        },
    }, //-- methods
    mounted() {
        // console.log(this.devid, this.channel, this.day);
        // this.getRecords(true);
        // $(document).on("keydown", this.keyDown);
    },
    beforeDestroy() {
        // $(document).off('keydown', this.keyDown);
        $(window).off("beforeunload", this.beforeUnload);
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
                console.log(this.devid, this.channel, this.day);
                this.getRecords(true);
            }
        })
	}
};
</script>

<template>
    <div :class="['modal', {fade: fade}]" data-backdrop="static" data-disable="false" data-keyboard="true" tabindex="-1">
        <div :class="['modal-dialog', size]">
            <div class="modal-content">
                <div class="modal-header">
                    <button type="button" class="close" data-dismiss="modal" aria-label="Close">
                        <span aria-hidden="true">&times;</span>
                    </button>
                    <h4 class="modal-title text-center text-primary"><span>{{title}}</span></h4>
                </div>
                <div class="modal-body">
                    <el-tabs v-model="tabName" @tab-click="tabClick">
                        <el-tab-pane name="status" v-loading="bStatusLogLoading" element-loading-text="加载中...">
                            <span slot="label"><i class="fa fa-list-ul"></i> 状态记录</span>
                            <el-timeline v-if="statusLogs.length > 0"
                                style="margin-top: 10px; margin-bottom: 10px; min-height: 200px; max-height: 450px; overflow-y:auto;">
                                <el-timeline-item v-for="(v, idx) in statusLogs" :key="idx" :type="statusTimeLineType(v)" :timestamp="v.CreatedAt">
                                    {{v.Description}}
                                </el-timeline-item>
                            </el-timeline>
                            <p v-else class="text-center">
                                <br>
                                <span v-if="bStatusLogLoading">暂无数据</span>
                                <span v-else-if="statusDays && statusDays > 0">最近{{statusDays}}天没有状态变更记录</span>
                                <span v-else>未找到最近的状态变更记录</span>
                            </p>
                        </el-tab-pane>
                        <el-tab-pane name="stream" v-loading="bStreamLogLoading" element-loading-text="加载中...">
                            <span slot="label"><i class="fa fa-bar-chart"></i> 流量统计</span>
                            <br>
                            <el-radio-group v-model="streamLogType" size="mini" @change="changeStreamLogType">
                                <el-radio-button label="today">今天</el-radio-button>
                                <el-radio-button label="yesterday">昨天</el-radio-button>
                                <el-radio-button label="week">最近7天</el-radio-button>
                                <el-radio-button label="month">最近30天</el-radio-button>
                            </el-radio-group>
                            <br>
                            <br>
                            <br>
                            <ve-line ref="streamChart" height="300px" :colors="chartColors" :data="streamLogs"
                                :extend="streamLogExtend" :settings="streamLogSettings" legend-position="bottom" :legend-visible="true"
                                :title="{text:'流量使用', left: 'center'}">
                            </ve-line>
                        </el-tab-pane>
                        <el-tab-pane name="position" v-loading="bPositionLogLoading" element-loading-text="加载中...">
                            <span slot="label"><i class="fa fa-location-arrow"></i> 移动轨迹</span>
                            <!-- <span slot="label"><i class="fa fa-map-marker"></i> 移动轨迹</span> -->
                            <el-timeline v-if="positionLogs.length > 0"
                                style="margin-top: 10px; margin-bottom: 10px; min-height: 200px; max-height: 450px; overflow-y:auto;">
                                <el-timeline-item v-for="(v, idx) in positionLogs" :key="idx" :type="positionTimeLineType(v)" :timestamp="v.CreatedAt">
                                    {{v.Method}} {{v.Longitude}}, {{v.Latitude}}
                                </el-timeline-item>
                            </el-timeline>
                            <p v-else class="text-center">
                                <br>
                                <span v-if="bPositionLogLoading">暂无数据</span>
                                <span v-else-if="positionDays && positionDays > 0">最近{{positionDays}}天没有位置变更记录</span>
                                <span v-else>未找到最近的位置变更记录</span>
                            </p>
                        </el-tab-pane>
                        <el-tab-pane name="session" v-if="serial && !code">
                            <span slot="label"><i class="fa fa-align-left"></i> 会话列表</span>
                            <div class="form-inline">
                                <div class="form-group form-group-sm">
                                    <label>搜索</label>
                                    <input type="text" class="form-control" placeholder="关键字" v-model.trim="sessionQ" @keydown.enter.prevent ref="sessionQ">
                                </div>
                                <span class="hidden-xs">&nbsp;&nbsp;</span>
                                <div class="form-group form-group-sm">
                                    <label>筛选</label>
                                    <select class="form-control" v-model.trim="sessionFilter">
                                        <option value="">全部</option>
                                        <option value="stream">直播</option>
                                        <option value="playback">回放</option>
                                        <option value="broadcast">对讲</option>
                                        <option value="record">录像</option>
                                        <option value="cascade">级联</option>
                                        <option value="playing">播放</option>
                                        <option value="udp">UDP</option>
                                        <option value="tcp">TCP</option>
                                        <option value="h264">H264</option>
                                        <option value="hevc">H265</option>
                                    </select>
                                </div>
                                <span class="hidden-xs" v-if="sessionNumOutputs && sessionFilter == 'playing'">&nbsp;&nbsp;</span>
                                <span v-if="sessionNumOutputs && sessionFilter == 'playing'">在线人数(合计): {{sessionNumOutputs}}</span>
                                <div class="form-group form-group-sm pull-right">
                                    <div class="btn-group btn-group-sm">
                                        <button type="button" class="btn btn-danger" @click.prevent="clearSessions" v-if="sessions.length && hasAnyRole(serverInfo, userInfo, '超级管理员')"><i class="fa fa-remove"></i> 清空</button>
                                        <button type="button" class="btn btn-primary" @click.prevent="getSessions"><i class="fa fa-refresh"></i> 刷新</button>
                                    </div>
                                </div>
                            </div>
                            <br>
                            <el-table :data="sessions" stripe :default-sort="{prop: 'StartAt', order: 'descending'}" @sort-change="sessionSortChange" :max-height="400" v-loading="sessionLoading" element-loading-text="加载中..." v-if="tabName == 'session'">
                                <!-- <el-table-column prop="DeviceID" label="设备国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column> -->
                                <el-table-column prop="ChannelID" label="通道国标编号" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                                <el-table-column prop="ChannelName" label="通道名称" min-width="120" :formatter="formatSessionChannelName" show-overflow-tooltip></el-table-column>
                                <el-table-column prop="SessionType" label="会话类型" min-width="100">
                                    <template slot-scope="props">
                                        <span :title="formatSessionTypeTitle(props.row)">{{formatSessionType(props.row)}}</span>
                                    </template>
                                </el-table-column>
                                <el-table-column prop="StartAt" label="开始时间" min-width="160" sortable="custom"></el-table-column>
                                <el-table-column prop="Duration" label="持续时长" min-width="120" :formatter="formatSessionDuration" align="right" sortable="custom"></el-table-column>
                                <el-table-column prop="InBytes" label="收流大小" min-width="120" :formatter="formatInBytes" align="right" show-overflow-tooltip sortable="custom"></el-table-column>
                                <el-table-column prop="InBitRate" label="收流码率" min-width="120" :formatter="formatInBitRate" align="right" show-overflow-tooltip sortable="custom"></el-table-column>
                                <el-table-column prop="Transport" label="流传输" min-width="100" show-overflow-tooltip></el-table-column>
                                <el-table-column prop="SourceVideoCodecName" label="视频编码" min-width="100" :formatter="formatVideoCodec"></el-table-column>
                                <el-table-column prop="CloudRecord" label="云端录像" min-width="100" :formatter="formatCloudRecord" align="center"></el-table-column>
                                <el-table-column prop="NumOutputs" label="在线人数" min-width="120" sortable="custom" :fixed="sessionFilter == 'playing' ? 'right' : false"></el-table-column>
                                <el-table-column prop="CascadeSize" label="级联数" min-width="100" sortable="custom"></el-table-column>
                                <el-table-column prop="SMSID" label="SMS" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                                <el-table-column label="操作" min-width="100" class-name="opt-group" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')">
                                    <template slot-scope="props">
                                        <div class="btn-group btn-group-xs">
                                            <button type="button" class="btn btn-danger" @click.prevent="stopSession(props.row)">
                                                <i class="fa fa-stop"></i> 停止
                                            </button>
                                        </div>
                                    </template>
                                </el-table-column>
                            </el-table>
                            <el-pagination v-if="sessionTotal > 0" layout="total,prev,pager,next,sizes" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="sessionTotal" :page-size.sync="sessionPageSize" :current-page.sync="sessionCurrentPage"></el-pagination>
                            <div class="clearfix"></div>
                        </el-tab-pane>
                    </el-tabs>
                </div>
                <div class="modal-footer" v-if="tabName == 'position' && serial && code && positionLogs.length > 0 && serverInfo.MapInfo && serverInfo.VersionType != '标准版'">
                    <a role="button" class="btn btn-primary" :href="`/map.html?serial=${serial}&code=${code}&trace=yes`" target="_blank">
                        <i class="fa fa-paper-plane"></i> 地图展示
                    </a>
                    <!-- <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button> -->
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import 'jquery-ui/ui/widgets/draggable';
    import $ from 'jquery';
    import _ from "lodash";
    import prettyBytes from 'pretty-bytes';
    import ElTimeline from 'components/Timeline.vue';
    import ElTimelineItem from 'components/TimelineItem.vue';
    import moment from 'moment';

    export default {
        props: {
            size: {
                type: String,
                default: ''
            },
            fade: {
                type: Boolean,
                default: false
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
        data() {
            return {
                tabName: "status",
                chartColors: ["#337ab7", "#7FFFD4"],
                streamLogSettings: {
                    area: true,
                    // xAxisType: "time",
                    labelMap: {
                        OutBytes:"发送",
                        InBytes:"接收",
                    },
                },
                streamLogExtend: {
                    legend: {
                        formatter: name => {
                            switch(name) {
                                case "接收":
                                    return `接收(${prettyBytes(this.streamLogs.sumInBytes)})`
                                    break;
                                case "发送":
                                    return `发送(${prettyBytes(this.streamLogs.sumOutBytes)})`
                                    break;
                            }
                            return name;
                        }
                    },
                    tooltip: {
                        formatter: arr => {
                            var name = arr[0].name;
                            if(this.streamLogType == 'today' || this.streamLogType == 'yesterday') {
                                name = moment(name, 'YYYY-MM-DD HH:mm:ss').format('HH');
                            } else {
                                name = moment(name, 'YYYY-MM-DD HH:mm:ss').format('MM-DD');
                            }
                            var tip = `${name}<br>${arr[0].seriesName} ${prettyBytes(arr[0].value[1])}`;
                            for(var i = 1; i < arr.length; i++) {
                                tip += `<br>${arr[i].seriesName} ${prettyBytes(arr[i].value[1])}`;
                            }
                            return tip;
                        },
                    },
                    yAxis: {
                        axisLabel: {
                            formatter: (name, idx) => {
                                return `${prettyBytes(name)}`;
                            }
                        }
                    },
                    xAxis: {
                        axisLabel: {
                            formatter: (name, idx) => {
                                if(this.streamLogType == 'today' || this.streamLogType == 'yesterday') {
                                    return `${moment(name, 'YYYY-MM-DD HH:mm:ss').format("HH")}`;
                                } else {
                                    return `${moment(name, 'YYYY-MM-DD HH:mm:ss').format('MM-DD')}`;
                                }
                            }
                        }
                    }
                },
                title: '状态记录',
                serial: '',
                code: '',
                limit: 20,

                statusLogs: [],
                statusDays: 0,
                bStatusLogLoading: false,

                positionLogs: [],
                positionDays: 0,
                bPositionLogLoading: false,

                streamLogType: "today",
                streamLogs: {
                    columns: ["Time", "OutBytes", "InBytes"],
                    rows: [],
                    sumInBytes: 0,
                    sumOutBytes: 0,
                },
                bStreamLogLoading: false,

                sessions: [],
                sessionNumOutputs: 0,
                sessionTotal: 0,
                sessionPageSize: 10,
                sessionCurrentPage: 1,
                sessionQ: "",
                sessionFilter: "",
                sessionSort: "",
                sessionOrder: "",
                sessionLoading: false,

                shown: false,
            }
        },
        watch: {
            sessionQ: function(val, oldVal) {
                this.delayGetSessions();
            },
            sessionFilter: function(val, oldVal) {
                this.getSessions();
            },
            sessionCurrentPage: function(val, oldVal) {
                this.getSessions();
            },
            sessionPageSize: function(val, oldVal) {
                this.getSessions();
            },
        },
        components: {
            ElTimeline, ElTimelineItem
        },
        mounted() {
            $(this.$el).find('.modal-content').draggable({
                handle: '.modal-header',
                cancel: ".modal-title span",
                addClasses: false,
                containment: 'document',
                delay: 100,
                opacity: 0.5
            });
            $(this.$el).on("shown.bs.modal", () => {
                this.shown = true;
                this.$emit("show");
            }).on("hidden.bs.modal", () => {
                this.shown = false;
                this.reset();
                this.$emit("hide");
            });
        },
        methods: {
            getStatusLogs() {
                if(!this.serial) return;
                this.bStatusLogLoading = true;
                this.statusLogs = [];
                $.get("/api/v1/device/statuslog", {
                    serial: this.serial,
                    code: this.code,
                    limit: this.limit,
                }).then(ret => {
                    if(ret) {
                        this.statusLogs = ret.LogList || [];
                        this.statusDays = ret.LogReserveDays || 0;
                    }
                }).always(() => {
                    this.$nextTick(() => {
                        this.bStatusLogLoading = false;
                    })
                })
            },
            getPositionLogs() {
                if(!this.serial) return;
                this.bPositionLogLoading = true;
                this.positionLogs = [];
                $.get("/api/v1/device/positionlog", {
                    serial: this.serial,
                    code: this.code,
                    limit: this.limit,
                }).then(ret => {
                    if(ret) {
                        this.positionLogs = ret.LogList || [];
                        this.positionDays = ret.LogReserveDays || 0;
                    }
                }).always(() => {
                    this.$nextTick(() => {
                        this.bPositionLogLoading = false;
                    })
                })
            },
            getStreamLogs() {
                if(!this.serial) return;
                this.bStreamLogLoading = true;
                this.streamLogs.rows = [];
                this.streamLogs.sumInBytes = 0;
                this.streamLogs.sumOutBytes = 0;
                $.get("/api/v1/device/streamlog", {
                    serial: this.serial,
                    code: this.code,
                    type: this.streamLogType,
                }).then(ret => {
                    if(ret) {
                        for(var v of ret.List) {
                            this.streamLogs.sumInBytes += v.InBytes;
                            this.streamLogs.sumOutBytes += v.OutBytes;
                        }
                        this.streamLogs.rows = ret.List || [];
                    }
                }).always(() => {
                    this.$nextTick(() => {
                        this.bStreamLogLoading = false;
                    })
                })
            },
            delayGetSessions: _.debounce(function() {
                this.getSessions();
            }, 800),
            getSessions(e) {
                if(!this.shown) return;
                this.sessionLoading = true;
                // this.sessions = [];
                $.get("/api/v1/device/session/list", {
                    serial: this.serial,
                    filter: this.sessionFilter,
                    q: this.sessionQ,
                    start: (this.sessionCurrentPage-1) * this.sessionPageSize,
                    limit: this.sessionPageSize,
                    sort: this.sessionSort,
                    order: this.sessionOrder,
                }).then(ret => {
                    if(ret) {
                        this.sessionNumOutputs = ret.NumOutputs || 0;
                        this.sessionTotal = ret.SessionCount || 0;
                        this.sessions = ret.SessionList || [];
                        if(e && e.target) {
                            this.$message({
                                type: "success",
                                message: "刷新成功！",
                            });
                        }
                    }
                }).always(() => {
                    this.$nextTick(() => {
                        this.sessionLoading = false;
                    })
                })
            },
            clearSessions() {
                $(this.$el).hide();
                this.$confirm(`确认清空 设备(${this.serial}) 会话?`, "提示", {
                    lockScroll: false,
                }).then(() => {
                    $(this.$el).show().focus();
                    $.post("/api/v1/device/session/clear", {
                        serial: this.serial,
                        filter: this.sessionFilter,
                        q: this.sessionQ,
                    }).always(() => {
                        this.getSessions();
                    });
                }).catch(() => {
                    $(this.$el).show().focus();
                });
            },
            stopSession(row) {
                $(this.$el).hide();
                this.$confirm(`确认停止 ${row.ChannelName || row.ChannelID} 会话?`, "提示", {
                    lockScroll: false,
                }).then(() => {
                    $(this.$el).show().focus();
                    $.post("/api/v1/device/session/stop", {
                        streamid: row.StreamID,
                    }).always(() => {
                        this.getSessions();
                    });
                }).catch(() => {
                    $(this.$el).show().focus();
                });
            },
            sessionSortChange(data) {
                this.sessionSort = data.prop;
                this.sessionOrder = data.order == "ascending" ? "asc" : "desc";
                this.getSessions();
            },
            formatSessionChannelName(row, col, cell) {
                return cell || "-";
            },
            formatSessionType(row, col, cell) {
                var strs = (row.StreamID || "").split(":");
                if(strs.length > 1) {
                    switch(strs[0]) {
                        case "stream":
                            return "直播";
                        case "playback":
                            return "回放";
                        case "download":
                            if(row.CloudRecord) {
                                return "补录";
                            } else {
                                return "下载";
                            }
                        case "broadcast":
                            return "对讲";
                        default:
                            return strs[0] || "-";
                    }
                }
                return "-";
            },
            formatSessionTypeTitle(row, col, cell) {
                if(row.PlaybackStartTime && row.PlaybackEndTime) {
                    return `${row.PlaybackStartTime} ~ ${row.PlaybackEndTime}`;
                }
                return "";
            },
            formatSessionDuration(row, col, cell) {
                return this.formatDuration(row.Duration || 0);
            },
            formatInBytes(row, col, cell) {
                return cell ? prettyBytes(cell) : "-";
            },
            formatInBitRate(row, col, cell) {
                return cell ? `${cell} Kbps` : "-";
            },
            formatVideoCodec(row, col, cell) {
                if(cell == "hevc") return "H265";
                if(cell == "h264") return "H264";
                return cell || "-";
            },
            formatCloudRecord(row, col, cell) {
                return cell ? "✓" : "✗";
            },
            statusTimeLineType(v) {
                var ret = "primary";
                if(!v) return ret;
                switch(v.Status) {
                    case "ON":
                        ret = "primary";
                        break;
                    case "OFF":
                        ret = "warning";
                        break;
                    case "DEL":
                        ret = "danger";
                        break;
                    default:
                        ret = "info";
                        break;
                }
                return ret;
            },
            positionTimeLineType(v) {
                var ret = "primary";
                if(!v) return ret;
                switch(v.Method) {
                    case "移动位置上报":
                    case "通道通知":
                    case "通道检索":
                        ret = "primary";
                        break;
                    case "接口设置":
                        ret = "warning";
                        break;
                    case "报警位置上报":
                        ret = "danger";
                        break;
                    default:
                        ret = "info";
                        break;
                }
                return ret;
            },
            tabClick(tab) {
                if(tab.name == 'stream') {
                    this.$nextTick(() => {
                        this.$refs["streamChart"].resize();
                    })
                }
            },
            changeStreamLogType() {
                this.getStreamLogs();
            },
            reset() {
                this.title = '状态记录';
                this.serial = '';
                this.code = '';

                this.statusLogs = [];
                this.statusDays = 0;

                this.positionLogs = [];
                this.positionDays = 0;

                this.streamLogs.rows = [];
                this.streamLogs.sumInBytes = 0;
                this.streamLogs.sumOutBytes = 0;
                this.streamLogType = "today";
                // this.tabName = "status";
                this.limit = 20;

                this.sessionQ = "";
                this.sessionFilter = "";
                this.sessionSort = "";
                this.sessionOrder = "";
                this.sessions = [];
                this.sessionCurrentPage = 1;
                this.sessionTotal = 0;
                this.sessionNumOutputs = 0;
                this.sessionPageSize = 10;
            },
            show(title, serial, code, limit) {
                this.title = title || this.title;
                this.serial = serial || '';
                this.code = code || '';
                this.limit = limit || 20;
                $(this.$el).modal("show");
                this.getStatusLogs();
                this.getPositionLogs();
                this.getStreamLogs();
            },
            showStatusLog(title, serial, code, limit) {
                this.tabName = "status";
                this.title = title || this.title;
                this.serial = serial || '';
                this.code = code || '';
                this.limit = limit || 20;
                $(this.$el).modal("show");
                this.getStatusLogs();
                this.getPositionLogs();
                this.getStreamLogs();
            },
            showStreamLog(title, serial, code) {
                this.tabName = "stream";
                this.title = title || this.title;
                this.serial = serial || '';
                this.code = code || '';
                $(this.$el).modal("show");
                this.getStatusLogs();
                this.getPositionLogs();
                this.getStreamLogs();
                this.$nextTick(() => {
                    this.$refs["streamChart"].resize();
                })
            },
            hide() {
                $(this.$el).modal("hide");
            }
        }
    }
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px){
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px){
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }
</style>

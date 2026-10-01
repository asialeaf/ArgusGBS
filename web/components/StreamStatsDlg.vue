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
                <div class="modal-body" v-loading="bStreamLogLoading" element-loading-text="加载中...">
                    <el-radio-group v-model="streamLogType" size="mini" @change="changeStreamLogType">
                        <el-radio-button label="now">当前</el-radio-button>
                        <el-radio-button label="today">今天</el-radio-button>
                        <el-radio-button label="yesterday">昨天</el-radio-button>
                        <el-radio-button label="week">最近7天</el-radio-button>
                        <el-radio-button label="month">最近30天</el-radio-button>
                    </el-radio-group>
                    <br>
                    <br>
                    <br>
                    <div style="text-align:center" v-if="streamLogType == 'now'">
                        <el-progress color="#337ab7" type="circle" :percentage="streamTotal == 0 ? 0 : parseFloat((streamOnline/streamTotal*100).toFixed(2))"></el-progress>
                        <div class="now-title">开流在线:<span class="now-num">{{streamOnline}}</span><br>开流总数:<span class="now-num">{{streamTotal}}</span></div>
                    </div>
                    <ve-line ref="streamChart" height="300px" :colors="chartColors" :data="streamLogs"
                        :extend="streamLogExtend" :settings="streamLogSettings" legend-position="bottom" :legend-visible="true" v-else>
                    </ve-line>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import 'jquery-ui/ui/widgets/draggable';
    import $ from 'jquery';
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
            }
        },
        data() {
            return {
                chartColors: ["#337ab7", "#7FFFD4"],
                streamLogSettings: {
                    area: true,
                    // xAxisType: "time",
                    labelMap: {
                        Online:"开流在线",
                        Total:"开流总数",
                    },
                },
                streamLogExtend: {
                    tooltip: {
                        formatter: arr => {
                            var name = arr[0].name;
                            if(this.streamLogType == 'today' || this.streamLogType == 'yesterday') {
                                name = moment(name, 'YYYY-MM-DD HH:mm:ss').format('HH');
                            } else {
                                name = moment(name, 'YYYY-MM-DD HH:mm:ss').format('MM-DD');
                            }
                            var tip = `${name}<br>${arr[0].seriesName} ${arr[0].value[1]}`;
                            for(var i = 1; i < arr.length; i++) {
                                tip += `<br>${arr[i].seriesName} ${arr[i].value[1]}`;
                            }
                            return tip;
                        },
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
                title: '开流统计',
                serial: '',
                code: '',

                streamLogType: "now",
                streamLogs: {
                    columns: ["Time", "Online", "Total"],
                    rows: [],
                },
                streamOnline: 0,
                streamTotal: 0,
                bStreamLogLoading: false,
                shown: false,
            }
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
            getStreamLogs() {
                this.bStreamLogLoading = true;
                this.streamLogs.rows = [];
                $.get("/api/v1/device/streamstats", {
                    serial: this.serial,
                    code: this.code,
                    type: this.streamLogType,
                }).then(ret => {
                    if(ret) {
                        if(this.streamLogType == "now" && ret.List && ret.List.length > 0) {
                            this.streamOnline = ret.List[0].Online || 0;
                            this.streamTotal = ret.List[0].Total || 0;
                        } else {
                            this.streamLogs.rows = ret.List || [];
                        }
                    }
                }).always(() => {
                    this.bStreamLogLoading = false;
                })
            },
            changeStreamLogType() {
                this.getStreamLogs();
            },
            reset() {
                this.title = '开流统计';
                this.serial = '';
                this.code = '';

                this.streamLogs.rows = [];
                this.streamOnline = 0;
                this.streamTotal = 0;
                this.streamLogType = "now";
            },
            show(title, serial, code) {
                this.title = title || this.title;
                this.serial = serial || '';
                this.code = code || '';
                $(this.$el).modal("show");
                this.getStreamLogs();
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

    .now-title {
        font-size: 12px;
        color: #1d3b55;
        margin-bottom: 15px;
    }

    .now-num {
        margin-top: 20%;
        font-size: 14px;
        font-weight: bold;
        color: #337ab7;
    }
</style>

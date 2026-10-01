<template>
    <div>
        <div class="box box-primary">
            <div class="box-header">
                <h4 class="text-primary text-center">报警列表</h4>
            </div>
            <div class="box-body">
                <div class="form-inline" autocomplete="off" spellcheck="false" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')">
                    <div class="form-group form-group-sm" style="display: inline-block;">
                        <div class="input-group input-group-sm">
                            <router-link :to="`/alarms/plans/1`" class="btn btn-sm btn-primary">
                                <i class="fa fa-calendar-check-o"></i> 报警预案
                            </router-link>
                        </div>
                    </div>
                    <div class="form-group form-group-sm pull-right" style="display: inline-block;">
                        <div class="input-group input-group-sm">
                            <button type="button" class="btn btn-sm btn-danger" @click.prevent="clearAlarm">
                                <i class="fa fa-remove"></i> 清空
                            </button>
                        </div>
                    </div>
                </div>
                <br class="hidden-xs" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')">
                <div class="clearfix" v-if="hasAnyRole(serverInfo, userInfo, '超级管理员')"></div>
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group form-group-sm">
                        <label>搜索</label>
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>报警级别</label>
                        <select class="form-control" v-model.trim="priority">
                            <option value="">全部</option>
                            <option value="1">一级警情</option>
                            <option value="2">二级警情</option>
                            <option value="3">三级警情</option>
                            <option value="4">四级警情</option>
                        </select>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>报警方式</label>
                        <select class="form-control" v-model.trim="method">
                            <option value="">全部</option>
                            <option value="1">电话报警</option>
                            <option value="2">设备报警</option>
                            <option value="3">短信报警</option>
                            <option value="4">GPS报警</option>
                            <option value="5">视频报警</option>
                            <option value="6">设备故障报警</option>
                            <option value="7">其他报警</option>
                        </select>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group">
                        <div class="input-group input-group-sm">
                            <DatePicker class="form-control input-group-sm" clearBtn :day.sync="starttime" ref="startDatePicker" placeholder="开始日期"></DatePicker>
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-default" @click.prevent="showStartDatePicker">
                                    <i class="fa fa-calendar"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                    <span class="hidden-xs">&nbsp;-&nbsp;</span>
                    <div class="form-group">
                        <div class="input-group input-group-sm">
                            <DatePicker class="form-control input-group-sm" clearBtn :day.sync="endtime" ref="endDatePicker" placeholder="结束日期"></DatePicker>
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-default" @click.prevent="showEndDatePicker">
                                    <i class="fa fa-calendar"></i>
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
                <br>
                <div class="clearfix"></div>
                <el-table :data="alarms" stripe :default-sort="{prop: 'Time', order: 'descending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column prop="DeviceID" label="报警设备" min-width="200" :formatter="formatDevice" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="ChannelID" label="报警通道" min-width="200" :formatter="formatChannel" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="ChannelName" label="通道名称" min-width="120" show-overflow-tooltip></el-table-column>
                    <el-table-column label="操作" min-width="100" :fixed="isMobile() ? false : 'right'">
                        <template slot-scope="props">
                            <div class="btn-group btn-group-xs">
                                <button type="button" class="btn btn-danger" @click.prevent="removeAlarm(props.row)" v-if="hasAnyRole(serverInfo, userInfo, '管理员')">
                                    <i class="fa fa-remove"></i> 删除
                                </button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column min-width="100" label="关联快照">
                        <template slot-scope="props">
                            <span v-if="!props.row.SnapPath" :title="props.row.PlanDescription">
                                <i>-</i>
                            </span>
                            <el-popover v-else :open-delay="1000" :close-delay="10" placement="left" :title="props.row.ChannelID" width="400" trigger="hover">
                                <img onerror="this.src='/images/default_snap.png';" style="width:100%;height:100%;" :src="props.row.SnapPath">
                                <img onerror="this.src='/images/default_snap.png';" style="height:30px;width:50px;" slot="reference" :src="props.row.SnapPath">
                            </el-popover>
                        </template>
                    </el-table-column>
                    <el-table-column min-width="200" label="关联录像" show-overflow-tooltip v-if="hasAnyRole(serverInfo, userInfo, '管理员', '超级管理员') && serverInfo.VersionType != '标准版'">
                        <template slot-scope="props">
                            <span v-if="!props.row.RecordPath" :title="props.row.PlanDescription">
                                <i>-</i>
                            </span>
                            <div class="btn-group btn-group-xs" v-else>
                                <button type="button" class="btn btn-primary" title="点击到新选项卡播放" @click.prevent="playRecord(props.row.RecordPath)" ><i class="fa fa-play-circle"> 播放</i></button>
                                <button type="button" class="btn btn-warning" title="点击到新选项卡下载" @click.prevent="downloadRecord(props.row.RecordPath)" ><i class="fa fa-download"> 下载</i></button>
                                <button type="button" class="btn btn-info" title="点击拷贝链接到粘贴板" v-clipboard="props.row.RecordPath" @success="$message({type:'success', message:'成功拷贝到粘贴板'})" ><i class="fa fa-copy"></i> 链接</button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column prop="AlarmPriority" label="报警级别" min-width="100" :formatter="formatPriority" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column prop="AlarmMethod" label="报警方式" min-width="100" :formatter="formatMethod" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="AlarmType" label="报警类型" min-width="140" :formatter="formatType" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Time" label="报警时间" min-width="160" sortable="custom">
                        <template slot-scope="props">
                            <span class="text-orange" v-if="!props.row.Time || props.row.Time == '0001-01-01 00:00:00'">{{props.row.CreatedAt}}</span>
                            <span v-else>{{props.row.Time}}</span>
                        </template>
                    </el-table-column>
                    <template slot="empty">
                        暂无数据, 请在 <router-link to="/devices">国标设备</router-link> 》 编辑 》 订阅项目 》 开启 报警订阅
                    </template>
                </el-table>
            </div>
            <div class="box-footer" v-if="total > 0">
                <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
            </div>
        </div>
        <div class="alert text-center" v-if="alarmPublishToRedis && serverInfo.DemoUser">
            <small>
                <strong><i class="fa fa-info-circle"></i> 提示 : </strong>
                已开启报警消息发布到 Redis, 可向 Redis 订阅 alarm 消息以获取实时报警 > SUBSCRIBE alarm 消息内容为报警信息 JSON 序列化字符串
            </small>
        </div>
    </div>
</template>

<script>
import _ from "lodash";
import DatePicker from "components/DatePicker.vue";
import { mapState } from "vuex";
import moment from 'moment';
export default {
    props: {},
    data() {
        return {
            q: "",
            priority: "",
            method: "",
            starttime: "",
            endtime: "",
            total: 0,
            pageSize: 10,
            currentPage: 1,
            sort: "Time",
            order: "desc",
            loading: false,
            timer: 0,
            alarms: [],
            alarmPublishToRedis: false,
        };
    },
    computed: {
        ...mapState(['userInfo', 'serverInfo']),
    },
    components: {
        DatePicker
    },
    mounted() {
        // this.timer = setInterval(() => {
        //   this.getAlarms();
        // }, 3000);
    },
    beforeDestroy() {
        if (this.timer) {
            clearInterval(this.timer);
            this.timer = 0;
        }
    },
    methods: {
        ready() {
            this.$watch('q', function(newVal, oldVal) {
                this.doDelaySearch();
            });
            this.$watch('starttime', function(newVal, oldVal) {
                this.doSearch();
            });
            this.$watch('endtime', function(newVal, oldVal) {
                this.doSearch();
            });
            this.$watch('priority', function(newVal, oldVal) {
                this.doSearch();
            });
            this.$watch('method', function(newVal, oldVal) {
                this.doSearch();
            });
            this.$watch('currentPage', function(newVal, oldVal) {
                this.doSearch(newVal);
            });
        },
        doSearch(page = 1) {
            var query = {};
            if (this.q) query["q"] = this.q;
            if (this.starttime) query["starttime"] = this.starttime;
            if (this.endtime) query["endtime"] = this.endtime;
            if (this.priority) query["priority"] = this.priority;
            if (this.method) query["method"] = this.method;
            this.$router.replace({
                path: `/alarms/${page}`,
                query: query
            });
        },
        doDelaySearch: _.debounce(function() {
            this.doSearch();
        }, 800),
        getAlarms(global = true) {
            if(global) {
                this.loading = true;
            }
            $.ajax({
                method: "GET",
                url: "/api/v1/alarm/list",
                global: global,
                data: {
                    q: this.q,
                    start: (this.currentPage -1) * this.pageSize,
                    limit: this.pageSize,
                    starttime: this.starttime ? moment(this.starttime, "YYYYMMDD").startOf('day').format("YYYY-MM-DDTHH:mm:ss") : "",
                    endtime: this.endtime ? moment(this.endtime, "YYYYMMDD").endOf('day').format("YYYY-MM-DDTHH:mm:ss") : "",
                    priority: this.priority,
                    method: this.method,
                    sort: this.sort,
                    order: this.order
                }
            }).then(ret => {
                this.total = ret.AlarmCount;
                this.alarms = ret.AlarmList;
                this.alarmPublishToRedis = ret.AlarmPublishToRedis;
            }).always(() => {
                if(global) {
                    this.$nextTick(() => {
                        this.loading = false;
                    })
                }
            });
        },
        sortChange(data) {
            this.sort = data.prop;
            this.order = data.order == "ascending" ? "asc" : "desc";
            this.getAlarms();
        },
        formatName(row, col, cell) {
            if (cell) return cell;
            return "-";
        },
        formatDevice(row, col, cell) {
            // if (row.DeviceName) {
            //   return row.DeviceName;
            // }
            if (row.DeviceID) {
                return row.DeviceID;
            }
            return "-";
        },
        formatChannel(row, rol, cell) {
            // if (row.ChannelName) {
            //   return row.ChannelName;
            // }
            if (row.ChannelID) {
                return row.ChannelID;
            }
            return "-";
        },
        formatPriority(row, col, cell) {
            return row.AlarmPriorityName || row.AlarmPriority || "-";
        },
        formatMethod(row, col, cell) {
            return row.AlarmMethodName || row.AlarmMethod || "-";
        },
        formatType(row, col, cell) {
            return row.AlarmTypeName || row.AlarmType || row.AlarmDescription || "-";
        },
        showStartDatePicker() {
            $(this.$refs.startDatePicker.$el).focus();
        },
        showEndDatePicker() {
            $(this.$refs.endDatePicker.$el).focus();
        },
        removeAlarm(row) {
            this.$confirm(`确认删除`, "提示", {
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/alarm/remove", {
                    id: row.ID
                }).always(() => {
                    this.getAlarms();
                });
            }).catch(() => {});
        },
        clearAlarm() {
            this.$confirm('确认全部删除', "提示", {
                lockScroll: false,
            }).then(() => {
                $.post("/api/v1/alarm/clear", {}).then(() => {
                    this.getAlarms();
                })
            }).catch(() => {});
        },
        downloadRecord(u) {
            window.open(u, "_blank");
        },
        playRecord(u) {
            window.open(u.replace("/api/v1/downloads/", "/api/v1/play/"), "_blank");
        }
    },
    beforeRouteEnter(to, from, next) {
        next(vm => {
            vm.q = to.query.q || "";
            vm.starttime = to.query.starttime || "";
            vm.endtime = to.query.endtime || "";
            vm.priority = to.query.priority || "";
            vm.method = to.query.method || "";
            vm.currentPage = parseInt(to.params.page) || 1;
            vm.ready();
        });
    },
    beforeRouteUpdate(to, from, next) {
        next();
        this.$nextTick(() => {
            this.q = to.query.q || "";
            this.starttime = to.query.starttime || "";
            this.endtime = to.query.endtime || "";
            this.priority = to.query.priority || "";
            this.method = to.query.method || "";
            this.currentPage = parseInt(to.params.page) || 1;
            this.alarms = [];
            this.total = 0;
            this.getAlarms();
        });
    }
};
</script>

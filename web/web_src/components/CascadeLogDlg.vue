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
                    <el-tabs v-model="tabName">
                        <el-tab-pane name="status" v-loading="bStatusLogLoading" element-loading-text="加载中...">
                            <span slot="label"><i class="fa fa-list-ul"></i> 状态记录</span>
                            <el-timeline v-if="statusLogs.length > 0"
                                style="margin-top: 10px; min-height: 200px; max-height: 435px; overflow-y:auto;">
                                <el-timeline-item v-for="(v, idx) in statusLogs" :key="idx" :type="statusTimeLineType(v)" :timestamp="v.CreatedAt">
                                    {{formatStatus(v)}}
                                    &nbsp;&nbsp;
                                    <span :class="['small', {'text-success': v.Status == 'ON', 'text-red': v.Status == 'OFF' && v.Description != 'CLOSED' && v.Description != 'STARTING'}]">{{v.Description}}</span>
                                </el-timeline-item>
                            </el-timeline>
                            <p v-else class="text-center">
                                <br>
                                <span v-if="bStatusLogLoading">暂无数据</span>
                                <span v-else-if="statusDays && statusDays > 0">最近{{statusDays}}天没有状态变更记录</span>
                                <span v-else>未找到最近的状态变更记录</span>
                            </p>
                        </el-tab-pane>
                    </el-tabs>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-default" data-dismiss="modal">关闭</button>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
    import 'jquery-ui/ui/widgets/draggable'
    import $ from 'jquery'
    import ElTimeline from 'components/Timeline.vue'
    import ElTimelineItem from 'components/TimelineItem.vue'

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
            limit: {
                type: Number,
                default: 20,
            },
        },
        data() {
            return {
                tabName: "status",
                title: '状态记录',
                cascadeInfo: null,

                statusLogs: [],
                statusDays: 0,
                bStatusLogLoading: false,
            }
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
                this.$emit("show");
            }).on("hidden.bs.modal", () => {
                this.reset();
                this.$emit("hide");
            });
        },
        methods: {
            getStatusLogs() {
                if(!this.cascadeInfo) return;
                this.bStatusLogLoading = true;
                this.statusLogs = [];
                $.get("/api/v1/cascade/statuslog", {
                    id: this.cascadeInfo.ID,
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
            formatStatus(v) {
                switch(v.Status) {
                    case "ON":
                        return "在线";
                    case "OFF":
                        return "离线";
                }
                return v.Status;
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
            reset() {
                this.title = '状态记录';
                this.cascadeInfo = null;

                this.statusLogs = [];
                this.statusDays = 0;
            },
            show(cascadeInfo) {
                this.title = `级联(${cascadeInfo.Name})`;
                this.cascadeInfo = cascadeInfo;
                $(this.$el).modal("show");
                this.getStatusLogs();
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

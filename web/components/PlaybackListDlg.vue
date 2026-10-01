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
                <el-table stripe :data="pageData" :default-sort="{prop: 'StartTime', order: 'ascending'}" @sort-change="sortChange" :max-height="500" @row-click="rowClick" @row-dblclick="rowDblclick" :row-style="rowStyle">
                    <el-table-column prop="DeviceID" label="通道国标编号" min-width="200" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Name" label="通道名称" min-width="120" :formatter="formatName" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="StartTime" label="开始时间" min-width="160" :formatter="formatName" sortable></el-table-column>
                    <el-table-column prop="EndTime" label="结束时间" min-width="160" :formatter="formatName" sortable></el-table-column>
                    <el-table-column prop="FileSize" label="文件大小" min-width="100" :formatter="formatFileSize" align="right" sortable></el-table-column>
                </el-table>
				<el-pagination v-if="total > 0" layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
				<div class="clearfix"></div>
			</div>
			<!-- <div class="modal-footer">
				<button type="button" class="btn btn-default" data-dismiss="modal">取消</button>
			</div> -->
		</div> <!-- /.modal-content -->
	</div> <!-- /.modal-dialog -->
</div>
</template>

<script>
import "jquery-ui/ui/widgets/draggable";
import prettyBytes from 'pretty-bytes';

export default {
	props: {
        title: {
            default: '录像列表'
        },
        size: {
            type: String,
            default: 'modal-lg'
        },
        fade: {
            type: Boolean,
            default: false
        },
        records: {
			type: Array,
			default: () => []
		}
	},
	data() {
		return {
			pageSize: 10,
			currentPage: 1,
			sort: "",
			order: "",
			selectable: false,
		};
	},
	computed: {
        total() {
            return this.records.length;
        },
        pageData() {
            let start = (this.currentPage - 1) * this.pageSize;
            let end = start + this.pageSize;
            return this.records.slice(start, end);
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
            this.$emit("show");
		}).on("hidden.bs.modal", () => {
            this.reset();
            this.$emit("hide");
		})
	},
	methods: {
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
        formatName(row, col, cell) {
            return cell || "-";
        },
        formatFileSize(row, col, cell) {
            return cell ? prettyBytes(cell) : "-";
        },
		rowClick(row, event, column) {
			if(this.selectable) {
				this.$emit("selected", row);
				this.hide();
			}
		},
		rowDblclick(row, event, column) {
			if(this.selectable) {
				this.$emit("selected", row);
				this.hide();
			}
		},
		rowStyle({row, rowIndex}) {
			if(this.selectable) {
				return "cursor:pointer";
			}
			return "";
		},
		reset() {
			this.selectable = false;
			this.currentPage = 1;
			this.pageSize = 10;
		},
		show(selectable = false) {
			this.selectable = selectable;
			$(this.$el).modal("show");
		},
		hide() {
			$(this.$el).modal("hide");
		},
	}, //-- methods
};
</script>

<style lang="less" scoped>
    .modal-content {
        overflow: hidden;
    }

    @media screen and(min-width: 992px) {
        .modal-dialog.modal-lgg {
            width: 90%;
        }
    }

    @media screen and(min-width: 1200px) {
        .modal-dialog.modal-lgg {
            width: 1200px;
        }
    }
</style>

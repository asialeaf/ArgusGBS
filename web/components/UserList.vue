<template>
    <div>
        <div class="box box-primary">
            <div class="box-header">
                <h4 class="text-primary text-center">用户列表</h4>
            </div>
            <div class="box-body">
                <div class="form-inline" autocomplete="off" spellcheck="false">
                    <div class="form-group form-group-sm">
                        <div class="input-group input-group-sm">
                            <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['userEditDlg'].show()">
                                <i class="fa fa-plus"></i> 添加用户
                            </button>
                        </div>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>搜索</label>
                        <div class="input-group input-group-sm" v-if="!isMobile() && hasAnyRole(serverInfo, userInfo, '超级管理员')">
                            <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q">
                            <div class="input-group-btn">
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="download" title="导出用户列表"><i class="fa fa-download"></i></button>
                                <button type="button" class="btn btn-sm btn-primary" @click.prevent="$refs['uploadDlg'].show()" title="导入用户列表"><i class="fa fa-upload"></i></button>
                            </div>
                        </div>
                        <input type="text" class="form-control" placeholder="关键字" v-model.trim="q" @keydown.enter.prevent ref="q" v-else>
                    </div>
                    <span class="hidden-xs">&nbsp;&nbsp;</span>
                    <div class="form-group form-group-sm">
                        <label>状态</label>
                        <select class="form-control" v-model.trim="enable">
                            <option value="">全部</option>
                            <option value="true">启用</option>
                            <option value="false">禁用</option>
                        </select>
                    </div>
                </div>
                <br>
                <div class="clearfix"></div>
                <el-table :data="users" stripe :default-sort="{prop: 'CreatedAt', order: 'descending'}" @sort-change="sortChange" v-loading="loading" element-loading-text="加载中...">
                    <el-table-column prop="Username" label="用户名" min-width="120" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column label="操作" min-width="280" :fixed="isMobile() ? false : 'right'" class-name="opt-group">
                        <template slot-scope="props">
                            <div class="btn-group btn-group-xs">
                                <button type="button" class="btn btn-warning" @click.prevent="editUser(props.row)">
                                    <i class="fa fa-edit"></i> 编辑
                                </button>
                                <button type="button" class="btn btn-primary" @click.prevent="editChannel(props.row)" v-if="!isReserve(props.row)">
                                    <i class="fa fa-check"></i> 关联通道
                                </button>
                                <button type="button" class="btn btn-info" @click.prevent="resetPassword(props.row)" v-if="!isReserve(props.row)">
                                    <i class="fa fa-key"></i> 重置密码
                                </button>
                                <button type="button" class="btn btn-danger" @click.prevent="removeUser(props.row)" v-if="!isReserve(props.row)">
                                    <i class="fa fa-remove"></i> 删除
                                </button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column prop="Role" label="角色" min-width="200" show-overflow-tooltip sortable="custom"></el-table-column>
                    <el-table-column prop="PhoneNumber" label="手机号" min-width="120" :formatter="formatPhoneNumber" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Email" label="邮箱" min-width="180" :formatter="formatEmail" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Description" label="备注" min-width="200" :formatter="formatDescription" show-overflow-tooltip></el-table-column>
                    <el-table-column prop="Enable" label="是否启用" min-width="100">
                        <template slot-scope="props">
                            <el-switch :value="props.row.Enable" @change="toggleEnable(props.row)" v-if="!isReserve(props.row)"></el-switch>
                        </template>
                    </el-table-column>
                    <el-table-column prop="LastLoginAt" label="最近登录" min-width="160" sortable="custom"></el-table-column>
                    <el-table-column prop="UpdatedAt" label="更新时间" min-width="160" sortable="custom"></el-table-column>
                    <el-table-column prop="CreatedAt" label="创建时间" min-width="160" sortable="custom"></el-table-column>
                </el-table>
            </div>
            <div class="box-footer" v-if="total > 0">
                <el-pagination layout="total,prev,pager,next" :pager-count="isMobile() ? 3 : 5" class="pull-right" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage"></el-pagination>
            </div>
        </div>
        <UploadDlg ref="uploadDlg" title="上传用户列表" url="/api/v1/user/import" @uploaded="uploaded"></UploadDlg>
        <UserEditDlg ref="userEditDlg" @submit="getUsers()"></UserEditDlg>
        <UserPasswordResetDlg ref="userPasswordResetDlg" :serverInfo="serverInfo"></UserPasswordResetDlg>
        <UserChannelListDlg ref="userChannelListDlg" :serverInfo="serverInfo" :userInfo="userInfo" :title="userChannelListDlgTitle"></UserChannelListDlg>
    </div>
</template>

<script>
import _ from "lodash";
import UploadDlg from "components/UploadDlg.vue";
import UserEditDlg from "components/UserEditDlg.vue";
import UserPasswordResetDlg from "components/UserPasswordResetDlg.vue";
import UserChannelListDlg from "components/UserChannelListDlg.vue";
import { mapState } from "vuex";
export default {
	data() {
		return {
			q: "",
			enable: "",
			total: 0,
			pageSize: 10,
			currentPage: 1,
			sort: "ID",
			order: "asc",
			loading: false,
			users: [],
			userChannelListDlgTitle: "关联通道",
		};
	},
	computed: {
		...mapState(['userInfo', 'serverInfo']),
	},
	components: {
		UploadDlg, UserEditDlg, UserPasswordResetDlg, UserChannelListDlg
	},
	mounted() {
		// this.timer = setInterval(() => {
		//   this.getUsers();
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
			this.$watch('enable', function(newVal, oldVal) {
				this.doSearch();
			});
			this.$watch('currentPage', function(newVal, oldVal) {
				this.doSearch(newVal);
			});
			// this.getUsers();
		},
		doSearch(page = 1) {
			var query = {};
			if (this.q) query["q"] = this.q;
			if (this.enable) query["enable"] = this.enable;
			this.$router.replace({
				path: `/user/${page}`,
				query: query
			});
		},
		doDelaySearch: _.debounce(function() {
			this.doSearch();
		}, 800),
		getUsers(global = true) {
			if(global) {
				this.loading = true;
			}
			$.ajax({
				method: "GET",
				url: "/api/v1/user/list",
				global: global,
				data: {
					q: this.q,
					start: (this.currentPage -1) * this.pageSize,
					limit: this.pageSize,
					enable: this.enable,
					sort: this.sort,
					order: this.order
				}
			}).then(ret => {
				this.total = ret.UserCount;
				this.users = ret.UserList;
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
			this.getUsers();
		},
		formatPhoneNumber(row, col, cell) {
			if (cell) return cell;
			return "-";
		},
		formatEmail(row, col, cell) {
			if (cell) return cell;
			return "-";
		},
		formatDescription(row, col, cell) {
			if (cell) return cell;
			return "-";
		},
		isReserve(row) {
			return row.Role == '超级管理员' && !row.Cas && !row.OAuth;
		},
		editUser(row) {
			this.$refs['userEditDlg'].show(row, this.isReserve(row));
		},
		editChannel(row) {
			this.userChannelListDlgTitle = `关联通道(${row.Username || row.ID})`;
			this.$refs['userChannelListDlg'].show(row.ID);
		},
		resetPassword(row) {
			this.$refs['userPasswordResetDlg'].show(row.ID, row.Username);
		},
		removeUser(row) {
			this.$confirm(`确认删除 ${row.Username || row.ID}`, "提示", {
				lockScroll: false,
			}).then(() => {
				$.post("/api/v1/user/remove", {
					id: row.ID
				}).always(() => {
					this.getUsers();
				});
			}).catch(() => {});
		},
		download() {
			window.open(`/api/v1/user/export?q=${this.q}&enable=${this.enable}`);
		},
		uploaded() {
				this.$message({
						type: 'success',
						message: "上传成功！"
				})
				this.getUsers();
				this.$refs['uploadDlg'].hide();
		},
		toggleEnable(row) {
			$.post("/api/v1/user/setenable", {
				id: row.ID,
				enable: !row.Enable,
			}).then(() => {
				row.Enable = !row.Enable;
			})
		},
	}, //-- methods
	beforeRouteEnter(to, from, next) {
		next(vm => {
			vm.q = to.query.q || "";
			vm.enable = to.query.enable || "";
			vm.currentPage = parseInt(to.params.page) || 1;
			vm.ready();
		});
	},
	beforeRouteUpdate(to, from, next) {
		next();
		this.$nextTick(() => {
			this.q = to.query.q || "";
			this.enable = to.query.enable || "";
			this.currentPage = parseInt(to.params.page) || 1;
			this.users = [];
			this.total = 0;
			this.getUsers();
		});
	}
};
</script>

<style lang="less">
.opt-group .cell {
	overflow: visible;
}
</style>

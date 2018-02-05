package sandbox

type TemplateDesc struct {
	Description string
	Notes string
	Contents string
}

type TemplateCollection  map[string]TemplateDesc
type AllTemplateCollection  map[string]TemplateCollection

// templates for single sandbox

var (
	Copyright string = `
#    DBDeployer - The MySQL Sandbox 
#    Copyright (C) 2006-2018 Giuseppe Maxia
#
#    Licensed under the Apache License, Version 2.0 (the "License");
#    you may not use this file except in compliance with the License.
#    You may obtain a copy of the License at
#                                                                             
#        http://www.apache.org/licenses/LICENSE-2.0
#                                                                             
#    Unless required by applicable law or agreed to in writing, software
#    distributed under the License is distributed on an "AS IS" BASIS,
#    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#    See the License for the specific language governing permissions and
#    limitations under the License.

# (Script generated by dbdeployer)

`
	init_db_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}
		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		SBDIR={{.SandboxDir}}
		cd $SBDIR
		if [ -d ./data/mysql ]
		then
			echo "Initialization already done."
			echo "This script should run only once."
			exit 0
		fi
		{{.InitScript}}
`

	start_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}
		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		MYSQLD_SAFE="bin/mysqld_safe"
		SBDIR={{.SandboxDir}}
		PIDFILE=$SBDIR/data/mysql_sandbox{{.Port}}.pid
		if [ ! -f $BASEDIR/$MYSQLD_SAFE ]
		then
			echo "mysqld_safe not found in $BASEDIR/bin/"
			exit 1
		fi
		MYSQLD_SAFE_OK=$(sh -n $BASEDIR/$MYSQLD_SAFE 2>&1)
		if [ "$MYSQLD_SAFE_OK" == "" ]
		then
			if [ "$SBDEBUG" == "2" ] 
			then
				echo "$MYSQLD_SAFE OK"
			fi
		else
			echo "$MYSQLD_SAFE has errors"
			echo "((( $MYSQLD_SAFE_OK )))"
			exit 1
		fi

		function is_running
		{
			if [ -f $PIDFILE ]
			then
				MYPID=$(cat $PIDFILE)
				ps -p $MYPID | grep $MYPID
			fi
		}

		TIMEOUT=180
		if [ -n "$(is_running)" ]
		then
			echo "sandbox server already started (found pid file $PIDFILE)"
		else
			if [ -f $PIDFILE ]
			then
				# Server is not running. Removing stale pid-file
				rm -f $PIDFILE
			fi
			CURDIR=$(pwd)
			cd $BASEDIR
			if [ "$SBDEBUG" = "" ]
			then
				$MYSQLD_SAFE --defaults-file=$SBDIR/my.sandbox.cnf $CUSTOM_MYSQLD $@ > /dev/null 2>&1 &
			else
				$MYSQLD_SAFE --defaults-file=$SBDIR/my.sandbox.cnf $CUSTOM_MYSQLD $@ > "$SBDIR/start.log" 2>&1 &
			fi
			cd $CURDIR
			ATTEMPTS=1
			while [ ! -f $PIDFILE ] 
			do
				ATTEMPTS=$(( $ATTEMPTS + 1 ))
				echo -n "."
				if [ $ATTEMPTS = $TIMEOUT ]
				then
					break
				fi
				sleep 1
			done
		fi

		if [ -f $PIDFILE ]
		then
			echo " sandbox server started"
		else
			echo " sandbox server not started yet"
			exit 1
		fi
`
	use_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}
		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		SBDIR={{.SandboxDir}}
		PIDFILE=$SBDIR/data/mysql_sandbox{{.Port}}.pid
		[ -n "$TEST_REPL_DELAY" -a -f $SBDIR/data/mysql-relay.index ] && sleep $TEST_REPL_DELAY
		[ -z "$MYSQL_EDITOR" ] && MYSQL_EDITOR="$BASEDIR/bin/mysql"
		if [ ! -x $MYSQL_EDITOR ]
		then
			if [ -x $SBDIR/$MYSQL_EDITOR ]
			then
				MYSQL_EDITOR=$SBDIR/$MYSQL_EDITOR
			else
				echo "MYSQL_EDITOR '$MYSQL_EDITOR' not found or not executable"
				exit 1
			fi
		fi
		HISTDIR=
		[ -z "$HISTDIR" ] && HISTDIR=$SBDIR
		export MYSQL_HISTFILE="$HISTDIR/.mysql_history"
		MY_CNF=$SBDIR/my.sandbox.cnf
		MY_CNF_NO_PASSWORD=$SBDIR/my.sandbox_np.cnf
		if [ -n "$NOPASSWORD" ]
		then
			grep -v '^password' < $MY_CNF > $MY_CNF_NO_PASSWORD
			MY_CNF=$MY_CNF_NO_PASSWORD
		fi
		if [ -f $PIDFILE ]
		then
			$MYSQL_EDITOR --defaults-file=$MY_CNF $MYCLIENT_OPTIONS "$@"
		else
			exit 1
		fi
`
	stop_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}
		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		SBDIR={{.SandboxDir}}
		PIDFILE=$SBDIR/data/mysql_sandbox{{.Port}}.pid

		MYSQL_ADMIN="$BASEDIR/bin/mysqladmin"


		function is_running
		{
			if [ -f $PIDFILE ]
			then
				MYPID=$(cat $PIDFILE)
				ps -p $MYPID | grep $MYPID
			fi
		}

		if [ -n "$(is_running)" ]
		then
			if [ -f $SBDIR/data/master.info ]
			then
				echo "stop slave" | $SBDIR/use -u root
			fi
			# echo "$MYSQL_ADMIN --defaults-file=$SBDIR/my.sandbox.cnf $MYCLIENT_OPTIONS shutdown"
			$MYSQL_ADMIN --defaults-file=$SBDIR/my.sandbox.cnf $MYCLIENT_OPTIONS shutdown
			sleep 1
		else
			if [ -f $PIDFILE ]
			then
				rm -f $PIDFILE
			fi
		fi

		if [ -n "$(is_running)" ]
		then
			# use the send_kill script if the server is not responsive
			$SBDIR/send_kill
		fi
`
	clear_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}

		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		SBDIR={{.SandboxDir}}
		PIDFILE=$SBDIR/data/mysql_sandbox{{.Port}}.pid
		cd $SBDIR

		#
		# attempt to drop databases gracefully
		#

		function is_running
		{
			if [ -f $PIDFILE ]
			then
				MYPID=$(cat $PIDFILE)
				ps -p $MYPID | grep $MYPID
			fi
		}

		if [ -n "$(is_running)" ]
		then
			for D in $(echo "show databases " | ./use -B -N | grep -v "^mysql$" | grep -iv "^information_schema$" | grep -iv "^performance_schema" | grep -ivw "^sys") 
			do
				echo "set sql_mode=ansi_quotes;drop database \"$D\"" | ./use 
			done
			VERSION={{.Version}}
			is_slave=$(ls data | grep relay)
			if [ -n "$is_slave" ]
			then
				./use -e "stop slave; reset slave;"
			fi
			if [[ "$VERSION" > "5.1" ]]
			then
				for T in general_log slow_log plugin
				do
					exists_table=$(./use -e "show tables from mysql like '$T'")
					if [ -n "$exists_table" ]
					then
						./use -e "truncate mysql.$T"
					fi
				done
			fi
		fi

		is_master=$(ls data | grep 'mysql-bin')
		if [ -n "$is_master" ]
		then
			./use -e 'reset master'
		fi

		./stop
		rm -f data/$(hostname)*
		rm -f data/log.0*
		rm -f data/*.log

		#
		# remove all databases if any (up to 8.0)
		#
		if [[ "$VERSION" < "8.0" ]]
		then
			for D in $(ls -d data/*/ | grep -w -v mysql | grep -iv performance_schema | grep -ivw sys)
			do
				rm -rf $D
			done
			mkdir data/test
		fi
`

	my_cnf_template string = `
		{{.Copyright}}
		# Template : {{.TemplateName}}
[mysql]
prompt='{{.Prompt}} [\h] {\u} (\d) > '
#

[client]
user               = {{.DbUser}}
password           = {{.DbPassword}}
port               = {{.Port}}
socket             = {{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock

[mysqld]
user               = {{.OsUser}}
port               = {{.Port}}
socket             = {{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock
basedir            = {{.Basedir}}
datadir            = {{.Datadir}}
tmpdir             = {{.Tmpdir}}
pid-file           = {{.Datadir}}/mysql_sandbox{{.Port}}.pid
bind-address       = {{.BindAddress}}
log-error=msandbox.err
{{.ServerId}}
{{.ReplOptions}}
{{.GtidOptions}}

{{.ExtraOptions}}
`
	send_kill_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}

		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		SBDIR={{.SandboxDir}}
		PIDFILE=$SBDIR/data/mysql_sandbox{{.Port}}.pid

		TIMEOUT=30

		function is_running
		{
			if [ -f $PIDFILE ]
			then
				MYPID=$(cat $PIDFILE)
				ps -p $MYPID | grep $MYPID
			fi
		}

		if [ -n "$(is_running)" ]
		then
			MYPID=$(cat $PIDFILE)
			echo "Attempting normal termination --- kill -15 $MYPID"
			kill -15 $MYPID
			# give it a chance to exit peacefully
			ATTEMPTS=1
			while [ -f $PIDFILE ]
			do
				ATTEMPTS=$(( $ATTEMPTS + 1 ))
				if [ $ATTEMPTS = $TIMEOUT ]
				then
					break
				fi
				sleep 1
			done
			if [ -f $PIDFILE ]
			then
				echo "SERVER UNRESPONSIVE --- kill -9 $MYPID"
				kill -9 $MYPID
				rm -f $PIDFILE
			fi
		else
			# server not running - removing stale pid-file
			if [ -f $PIDFILE ]
			then
				rm -f $PIDFILE
			fi
		fi
`
	status_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}

		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		SBDIR={{.SandboxDir}}
		PIDFILE=$SBDIR/data/mysql_sandbox{{.Port}}.pid
		baredir=$(basename $SBDIR)

		source $SBDIR/sb_include
		node_status=off
		exit_code=1
		if [ -f $PIDFILE ]
		then
			MYPID=$(cat $PIDFILE)
			running=$(ps -p $MYPID | grep $MYPID)
			if [ -n "$running" ]
			then
				node_status=on
				exit_code=0
			fi
		fi
		echo "$baredir $node_status"
		exit $exit_code
`
	restart_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}

		SBDIR={{.SandboxDir}}
		$SBDIR/stop
		$SBDIR/start $@
`
	load_grants_template string = `#!/bin/bash
		{{.Copyright}}
		# Template : {{.TemplateName}}

		BASEDIR={{.Basedir}}
		export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
		export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
		SBDIR={{.SandboxDir}}
		PIDFILE=$SBDIR/data/mysql_sandbox{{.Port}}.pid

		source $SBDIR/sb_include
		MYSQL="$BASEDIR/bin/mysql --no-defaults --socket={{.GlobalTmpDir}}/mysql_sandbox{{.Port}}.sock --port={{.Port}}"
		VERBOSE_SQL=''
		[ -n "$SBDEBUG" ] && VERBOSE_SQL=-v
		$MYSQL -u root $VERBOSE_SQL < $SBDIR/grants.mysql
`
	grants_template5x string = `
		# Template : {{.TemplateName}}
use mysql;
set password=password('{{.DbPassword}}');
grant all on *.* to {{.DbUser}}@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
grant all on *.* to {{.DbUser}}@'localhost' identified by '{{.DbPassword}}';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
    on *.* to msandbox_rw@'localhost' identified by '{{.DbPassword}}';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
    on *.* to msandbox_rw@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
grant SELECT,EXECUTE on *.* to msandbox_ro@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
grant SELECT,EXECUTE on *.* to msandbox_ro@'localhost' identified by '{{.DbPassword}}';
grant REPLICATION SLAVE on *.* to {{.RplUser}}@'{{.RemoteAccess}}' identified by '{{.RplPassword}}';
delete from user where password='';
delete from db where user='';
flush privileges;
create database if not exists test;
`
	grants_template57 string = `

		# Template : {{.TemplateName}}
use mysql;
set password='{{.DbPassword}}';
-- delete from tables_priv;
-- delete from columns_priv;
-- delete from db;
-- delete from user where user not in ('root', 'mysql.sys', 'mysqlxsys', 'mysql.session', 'mysql.infoschema');
-- flush privileges;

create user {{.DbUser}}@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
grant all on *.* to {{.DbUser}}@'{{.RemoteAccess}}' ;

create user {{.DbUser}}@'localhost' identified by '{{.DbPassword}}';
grant all on *.* to {{.DbUser}}@'localhost';

create user msandbox_rw@'localhost' identified by '{{.DbPassword}}';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
    on *.* to msandbox_rw@'localhost';

create user msandbox_rw@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
grant SELECT,INSERT,UPDATE,DELETE,CREATE,DROP,INDEX,ALTER,
    SHOW DATABASES,CREATE TEMPORARY TABLES,LOCK TABLES, EXECUTE 
    on *.* to msandbox_rw@'{{.RemoteAccess}}';

create user msandbox_ro@'{{.RemoteAccess}}' identified by '{{.DbPassword}}';
create user msandbox_ro@'localhost' identified by '{{.DbPassword}}';
create user {{.RplUser}}@'{{.RemoteAccess}}' identified by '{{.RplPassword}}';
grant SELECT,EXECUTE on *.* to msandbox_ro@'{{.RemoteAccess}}';
grant SELECT,EXECUTE on *.* to msandbox_ro@'localhost';
grant REPLICATION SLAVE on *.* to {{.RplUser}}@'{{.RemoteAccess}}';
create schema if not exists test;
`
	add_option_template string = `#!/bin/bash
{{.Copyright}}
# Template : {{.TemplateName}}

curdir="{{.SandboxDir}}"
cd $curdir

if [ -z "$*" ]
then
    echo "# Syntax $0 options-for-my.cnf [more options] "
    exit
fi

CHANGED=''
for OPTION in $@
do
    option_exists=$(grep $OPTION ./my.sandbox.cnf)
    if [ -z "$option_exists" ]
    then
        echo "$OPTION" >> my.sandbox.cnf
        echo "# option '$OPTION' added to configuration file"
        CHANGED=1
    else
        echo "# option '$OPTION' already exists configuration file"
    fi
done

if [ -n "$CHANGED" ]
then
    ./restart
fi
`
	show_binlog_template string = `#!/bin/bash
{{.Copyright}}
# Template : {{.TemplateName}}

curdir="{{.SandboxDir}}"
cd $curdir

if [ ! -d ./data ]
then
    echo "$curdir/data not found"
    exit 1
fi

pattern=$1
[ -z "$pattern" ] && pattern='[0-9]*'
if [ "$pattern" == "-h" -o "$pattern" == "--help" -o "$pattern" == "-help" -o "$pattern" == "help" ]
then
    echo "# Usage: $0 [BINLOG_PATTERN] "
    echo "# Where BINLOG_PATTERN is a number, or part of a number used after 'mysql-bin'"
    echo "# (The default is '[0-9]*]')"
    echo "# examples:" 
    echo "#          ./show_binlog 000001 | less "
    echo "#          ./show_binlog 000012 | vim - "
    echo "#          ./show_binlog  | grep -i 'CREATE TABLE'"
    exit 0
fi
# set -x
last_binlog=$(ls -lotr data/mysql-bin.$pattern | tail -n 1 | awk '{print $NF}')

if [ -z "$last_binlog" ]
then
    echo "No binlog found in $curdir/data"
    exit 1
fi

# Checks if the output is a terminal or a pipe
if [  -t 1 ]
then
    echo "###################### WARNING ####################################"
    echo "# You are not using a pager."
    echo "# The output of this script can be quite large."
    echo "# Please pipe this script with a pager, such as 'less' or 'vim -'"
    echo "# ENTER 'q' to exit or simply RETURN to continue without a pager"
    read answer
    if [ "$answer" == "q" ]
    then
        exit
    fi
fi

(printf "#\n# Showing $last_binlog\n#\n" ; ./my sqlbinlog --verbose $last_binlog ) 
`
	my_template string = `#!/bin/bash
{{.Copyright}}
# Template : {{.TemplateName}}


if [ "$1" = "" ]
then
    echo "syntax my sql{dump|binlog|admin} arguments"
    exit
fi

SBDIR="{{.SandboxDir}}"
source $SBDIR/sb_include
BASEDIR={{.Basedir}}
export LD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH=$BASEDIR/lib:$BASEDIR/lib/mysql:$DYLD_LIBRARY_PATH
MYSQL=$BASEDIR/bin/mysql

SUFFIX=$1
shift

MYSQLCMD="$BASEDIR/bin/my$SUFFIX"

NODEFAULT=(myisam_ftdump
myisamlog
mysql_config
mysql_convert_table_format
mysql_find_rows
mysql_fix_extensions
mysql_fix_privilege_tables
mysql_secure_installation
mysql_setpermission
mysql_tzinfo_to_sql
mysql_config_editor
mysql_waitpid
mysql_zap
mysqlaccess
mysqlbinlog
mysqlbug
mysqldumpslow
mysqlhotcopy
mysqltest
mysqltest_embedded)

DEFAULTSFILE="--defaults-file=$SBDIR/my.sandbox.cnf"

for NAME in ${NODEFAULT[@]}
do
    if [ "my$SUFFIX" = "$NAME" ]
    then
        DEFAULTSFILE=""
        break
    fi
done

if [ -f $MYSQLCMD ]
then
    $MYSQLCMD $DEFAULTSFILE "$@"
else
    echo "$MYSQLCMD not found "
fi
`
	show_relaylog_template string=`#!/bin/bash
{{.Copyright}}
# Template : {{.TemplateName}}
curdir="{{.SandboxDir}}"
cd $curdir

if [ ! -d ./data ]
then
    echo "$curdir/data not found"
    exit 1
fi
relay_basename=$1
[ -z "$relay_basename" ] && relay_basename='mysql-relay'
pattern=$2
[ -z "$pattern" ] && pattern='[0-9]*'
if [ "$pattern" == "-h" -o "$pattern" == "--help" -o "$pattern" == "-help" -o "$pattern" == "help" ]
then
    echo "# Usage: $0 [ relay-base-name [BINLOG_PATTERN]] "
    echo "# Where relay-basename is the initial part of the relay ('$relay_basename')"
    echo "# and BINLOG_PATTERN is a number, or part of a number used after '$relay_basename'"
    echo "# (The default is '[0-9]*]')"
    echo "# examples:" 
    echo "#          ./show_relaylog relay-log-alpha 000001 | less "
    echo "#          ./show_relaylog relay-log 000012 | vim - "
    echo "#          ./show_relaylog  | grep -i 'CREATE TABLE'"
    exit 0
fi
# set -x
last_relaylog=$(ls -lotr data/$relay_basename.$pattern | tail -n 1 | awk '{print $NF}')

if [ -z "$last_relaylog" ]
then
    echo "No relay log found in $curdir/data"
    exit 1
fi

# Checks if the output is a terminal or a pipe
if [  -t 1 ]
then
    echo "###################### WARNING ####################################"
    echo "# You are not using a pager."
    echo "# The output of this script can be quite large."
    echo "# Please pipe this script with a pager, such as 'less' or 'vim -'"
    echo "# ENTER 'q' to exit or simply RETURN to continue without a pager"
    read answer
    if [ "$answer" == "q" ]
    then
        exit
    fi
fi

(printf "#\n# Showing $last_relaylog\n#\n" ; ./my sqlbinlog --verbose $last_relaylog ) 
`
	sb_include_template string = ""

SingleTemplates  = TemplateCollection{
	"Copyright" : TemplateDesc{
			Description: "Copyright for every sandbox script",
			Notes: "",
			Contents : Copyright,
		},
		"init_db_template" : TemplateDesc{
			Description : "Initialization template for the database",
			Notes : "This should normally run only once",
			Contents : init_db_template,
		},
		"start_template" : TemplateDesc{
			Description : "starts the database in a single sandbox (with optional mysqld arguments)",
			Notes : "",
			Contents : start_template,
		},
		"use_template" : TemplateDesc{
			Description : "Invokes the MySQL client with the appropriate options",
			Notes : "",
			Contents : use_template,
		},
		"stop_template" : TemplateDesc{
			Description : "Stops a database in a single sandbox",
			Notes : "",
			Contents : stop_template,
		},
		"clear_template" : TemplateDesc{
			Description : "Remove all data from a single sandbox",
			Notes : "",
			Contents : clear_template,
		},
		"my_cnf_template" : TemplateDesc{
			Description : "Default options file for a sandbox",
			Notes : "",
			Contents : my_cnf_template,
		},
		"status_template" : TemplateDesc{
			Description : "Shows the status of a single sandbox",
			Notes : "",
			Contents : status_template,
		},
		"restart_template" : TemplateDesc{
			Description : "Restarts the database (with optional mysqld arguments)",
			Notes : "",
			Contents : restart_template,
		},
		"send_kill_template" : TemplateDesc{
			Description : "Sends a kill signal to the database",
			Notes : "",
			Contents : send_kill_template,
		},
		"load_grants_template" : TemplateDesc{
			Description : "Loads the grants defined for the sandbox",
			Notes : "",
			Contents : load_grants_template,
		},
		"grants_template5x" : TemplateDesc{
			Description : "Grants for sandboxes up to 5.6",
			Notes : "",
			Contents : grants_template5x,
		},
		"grants_template57" : TemplateDesc{
			Description : "Grants for sandboxes from 5.7+",
			Notes : "",
			Contents : grants_template57,
		},
		"my_template" : TemplateDesc{
			Description : "Prefix script to run every my* command line tool",
			Notes : "",
			Contents : my_template,
		},
		"add_option_template" : TemplateDesc{
			Description : "Adds options to the my.sandbox.cnf file and restarts",
			Notes : "",
			Contents : add_option_template,
		},
		"show_binlog_template" : TemplateDesc{
			Description : "Shows a binlog for a single sandbox",
			Notes : "",
			Contents : show_binlog_template,
		},
		"show_relaylog_template" : TemplateDesc{
			Description : "Show the relaylog for a single sandbox",
			Notes : "",
			Contents : show_relaylog_template,
		},
		"sb_include_template" : TemplateDesc{
			Description : "TBD",
			Notes : "",
			Contents : sb_include_template,
		},
	}
	AllTemplates = AllTemplateCollection{
		"single" : SingleTemplates,
		"multiple" : MultipleTemplates,
		"replication" : ReplicationTemplates,
		"group" : GroupTemplates,
	}
)



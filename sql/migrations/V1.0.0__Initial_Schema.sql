create table report (
	id int not null auto_increment primary key,
	ua int not null,
	eventDT DATETIME not null,
	remoteIP int(4) not null,
	host char(255) not null,
	uriPath char(255) not null,
	documentURI char(255) not null,
	report blob not null,
	INDEX idx_documentURI (documentURI),
	INDEX idx_host (host),
	INDEX idx_host_path (host,uriPath)
	
) Engine=MyISAM;

create table userAgent (
	id int not null auto_increment primary key,
	crc32 int(4) not null,
	ua char(255) not null,
	INDEX idx_crc32 (crc32)
) Engine=MyISAM;

-- phpMyAdmin SQL Dump
-- version 4.2.8.1
-- http://www.phpmyadmin.net
--
-- Host: localhost
-- Generation Time: Feb 25, 2015 at 07:03 PM
-- Server version: 5.6.20
-- PHP Version: 5.5.14

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8 */;

--
-- Database: `imp`
--

-- --------------------------------------------------------

--
-- Table structure for table `Guest`
--

CREATE TABLE IF NOT EXISTS `Guest` (
`GuestId` int(11) NOT NULL,
  `Handle` varchar(16) NOT NULL,
  `Host` varchar(255) NOT NULL,
  `Token` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Table structure for table `HandleLimit`
--

CREATE TABLE IF NOT EXISTS `HandleLimit` (
  `Handle` varchar(16) NOT NULL,
  `LoginAttemptCount` int(11) NOT NULL,
  `LastAttemptDate` datetime NOT NULL,
  `NextLoginDelay` int(11) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Table structure for table `Host`
--

CREATE TABLE IF NOT EXISTS `Host` (
`HostId` int(11) NOT NULL,
  `Name` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Table structure for table `IPLimit`
--

CREATE TABLE IF NOT EXISTS `IPLimit` (
  `IP` varchar(45) NOT NULL,
  `LastLoginAttemptDate` datetime DEFAULT NULL,
  `UsersAllowedCount` tinyint(4) NOT NULL,
  `CountResetDate` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Table structure for table `User`
--

CREATE TABLE IF NOT EXISTS `User` (
`UserId` int(11) NOT NULL,
  `Handle` varchar(16) NOT NULL,
  `Status` varchar(140) NOT NULL DEFAULT '',
  `Biography` varchar(140) NOT NULL DEFAULT '',
  `Email` varchar(254) NOT NULL,
  `IsValidEmail` tinyint(1) NOT NULL DEFAULT '0',
  `EmailValidationToken` varchar(50) DEFAULT NULL,
  `EmailValidationDate` date DEFAULT NULL,
  `PasswordHash` varchar(60) NOT NULL,
  `JoinedDate` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `IsDisabled` tinyint(1) NOT NULL DEFAULT '0'
) ENGINE=InnoDB AUTO_INCREMENT=13 DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Table structure for table `UserHost`
--

CREATE TABLE IF NOT EXISTS `UserHost` (
  `UserId` int(11) NOT NULL,
  `HostId` int(11) NOT NULL,
  `Token` varchar(255) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

-- --------------------------------------------------------

--
-- Table structure for table `UserToken`
--

CREATE TABLE IF NOT EXISTS `UserToken` (
  `Token` varchar(50) NOT NULL,
  `UserId` int(11) NOT NULL,
  `LoginTime` date NOT NULL,
  `LastSeenTime` date NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=latin1;

--
-- Indexes for dumped tables
--

--
-- Indexes for table `Guest`
--
ALTER TABLE `Guest`
 ADD PRIMARY KEY (`GuestId`);

--
-- Indexes for table `HandleLimit`
--
ALTER TABLE `HandleLimit`
 ADD PRIMARY KEY (`Handle`);

--
-- Indexes for table `Host`
--
ALTER TABLE `Host`
 ADD PRIMARY KEY (`HostId`);

--
-- Indexes for table `IPLimit`
--
ALTER TABLE `IPLimit`
 ADD PRIMARY KEY (`IP`);

--
-- Indexes for table `User`
--
ALTER TABLE `User`
 ADD PRIMARY KEY (`UserId`), ADD UNIQUE KEY `Handle` (`Handle`), ADD UNIQUE KEY `Email` (`Email`);

--
-- Indexes for table `UserToken`
--
ALTER TABLE `UserToken`
 ADD PRIMARY KEY (`Token`);

--
-- AUTO_INCREMENT for dumped tables
--

--
-- AUTO_INCREMENT for table `Guest`
--
ALTER TABLE `Guest`
MODIFY `GuestId` int(11) NOT NULL AUTO_INCREMENT;
--
-- AUTO_INCREMENT for table `Host`
--
ALTER TABLE `Host`
MODIFY `HostId` int(11) NOT NULL AUTO_INCREMENT;
--
-- AUTO_INCREMENT for table `User`
--
ALTER TABLE `User`
MODIFY `UserId` int(11) NOT NULL AUTO_INCREMENT,AUTO_INCREMENT=3;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;

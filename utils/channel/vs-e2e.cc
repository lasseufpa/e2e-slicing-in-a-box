/* -*-  Mode: C++; c-file-style: "gnu"; indent-tabs-mode:nil; -*- */

// Copyright (c) 2021 Centre Tecnologic de Telecomunicacions de Catalunya (CTTC)
//
// SPDX-License-Identifier: GPL-2.0-only

/**
 * \ingroup scratch
 * \file vs-e2e.cc
 * \brief A cozy, simple, NR MIMO demo (in a tutorial style)
 *
 * This example describes how to setup a MIMO simulation using the 3GPP channel model
 * from TR 38.900. This example consists of a simple topology, in which there
 * is only one gNB and two UE's. Have a look at the possible parameters
 * to know what you can configure through the command line.
 *
 * The example will emulate the 5G channel. Using tap-bridge module, both containers 
 * can interact with each other.
 *
 * \code{.unparsed}
$ ./ns3 run "vs-e2e"
    \endcode
 *
 */

/*
 * Include part. Often, you will have to include the headers for an entire module;
 * do that by including the name of the module you need with the suffix "-module.h".
 */
#include "ns3/ipv4-global-routing-helper.h"
#include "ns3/antenna-module.h"
#include "ns3/applications-module.h"
#include "ns3/config-store-module.h"
#include "ns3/core-module.h"
#include "ns3/flow-monitor-module.h"
#include "ns3/internet-apps-module.h"
#include "ns3/internet-module.h"
#include "ns3/mobility-module.h"
#include "ns3/network-module.h"
#include "ns3/epc-helper.h"
#include "ns3/lte-module.h"
#include "ns3/nr-module.h"
#include "ns3/point-to-point-module.h"
#include "ns3/tap-bridge-module.h"
#include "ns3/csma-module.h"
#include <fstream>
#include <iostream>
#include "ns3/log.h"
#include "ns3/fd-net-device-module.h"
#include <errno.h>
#include <sys/socket.h>
/*
 * Use, always, the namespace ns3. All the NR classes are inside such namespace.
 */
using namespace ns3;

/*
 * With this line, we will be able to see the logs of the file by enabling the
 * component "vs-e2e"
 */
NS_LOG_COMPONENT_DEFINE("vs-e2e");



/**
 * \author @Vadym Hapanchak ~ https://github.com/vandit86/ns3-scratch
 * \brief addBackAdress() add a packet flow through pgw and the container
 * \param pgw a PacketGateway node
 * \param ueNrNetDev Ue container devices
 * \param addr Container address  
*/
void
addBackAddress (Ptr<Node> pgw, Ptr<NetDevice> ueNrNetDev, Ipv4Address addr)
{
  // First we need to get IMSI of ueNrNetDevice conencted through csma link to container
  Ptr<NrUeNetDevice> uedev = ueNrNetDev->GetObject<NrUeNetDevice> ();
  uint64_t imsi = uedev->GetImsi ();

  // get PGW application from pgw node
  Ptr<EpcPgwApplication> pgwApp = pgw->GetApplication (0)->GetObject<EpcPgwApplication> ();

  // add container address to allow traffic through PGW node
  pgwApp->SetUeAddress (imsi, addr);
}

/**
 * \brief MacAdress() is a simple function to print all UE's and gNB MacAddress;
 * \param enbnetdev is a NetDeviceContainer for all the gNB's;
 * \param uenetdev is a NetDeviceContainer for all the UE's;
 * 
*/
void MacAddress(NetDeviceContainer enbnetdev, NetDeviceContainer uenetdev){
    
    std::cout << "------------ MAC'S Address ------------" << std::endl;
    for (uint32_t i = 0; i<enbnetdev.GetN();i++)
    {
        Ptr<NetDevice> ueDevice = enbnetdev.Get(i);
        Ptr <NrNetDevice> nrnet = DynamicCast<NrNetDevice>(ueDevice);
        Address macAddress = nrnet->GetAddress();
        std::cout << "gNB Mac Address:"<< macAddress << std::endl;
    }
    for (uint32_t i = 0; i<uenetdev.GetN();i++)
    {
        Ptr<NetDevice> ueDevice = uenetdev.Get(i);
        Ptr <NrNetDevice> nrnet = DynamicCast<NrNetDevice>(ueDevice);
        Address macAddress = nrnet->GetAddress();
        std::cout << "UE"<< i << " Mac Address:" << macAddress << std::endl;
    }
    //std::cout << "---------------------------------------" << std::endl;
}

/**
 * \brief MacAdress() is a simple function to print all UE's and gNB MacAddress;
 * \param allnodes is a Nodecontainer that contains Nodes;
 * 
*/
void ipv4(NodeList allnodes){
    
    std::cout << "------------ Ipv4 Address ------------" << std::endl;
    for (uint32_t i = 0; i<allnodes.GetNNodes();i++)
    {
        Ptr <Node> node = allnodes.GetNode(i);
        Ptr<Ipv4> nodeIpv4 = node->GetObject<Ipv4> ();
        for(uint32_t Ni = 0; Ni < (node->GetObject<Ipv4>()->GetNInterfaces()); Ni++){
            Ipv4Address nodeAddr = nodeIpv4->GetAddress (Ni, 0).GetAddress ();
            std::cout << "Ip of "<< Names::FindName(allnodes.GetNode(i)) <<": "<< nodeAddr << std::endl;}
        std::cout << std::endl;
    }
}
/**
 * \brief PrintRoutingTable() small function to print all route rules
 * \param allnodes all nodes and ghost nodes used in the scenario
*/
void PrintRoutingTable(NodeList allnodes) {
    Ipv4StaticRoutingHelper ipv4RoutingHelper;
    std::cout << "----------------Routes----------------" << std::endl;
    for (uint32_t i = 0; i < allnodes.GetNNodes(); i++){
        Ptr <Node> node = allnodes.GetNode(i);
        Ptr <Ipv4> node_ipv4 = node->GetObject<Ipv4>();
        Ptr <Ipv4StaticRouting> staticRouting = ipv4RoutingHelper.GetStaticRouting (node_ipv4);
        uint32_t nRoutes = staticRouting->GetNRoutes ();
        std::cout << Names::FindName(node) << ":" << std::endl;
        std::cout << std::endl;
        for (uint32_t j = 0; j < nRoutes; j++)
        {
            std::cout << "Route " << j << std::endl;
            Ipv4RoutingTableEntry route = staticRouting->GetRoute (j);
            std::cout << "\tDestination: "<< route.GetDest() << " (" << route.GetDestNetworkMask() << ")" << std::endl;
            std::cout << "\tDefault Gateway: " << route.GetGateway() << std::endl;
            std::cout << "\tOutput Interface: " << route.GetInterface() << std::endl;
        }
        std::cout << "--------------------------------------------" << std::endl;
    }
}
int
main(int argc, char* argv[])
{
    GlobalValue::Bind("SimulatorImplementationType", StringValue("ns3::RealtimeSimulatorImpl"));
    GlobalValue::Bind("ChecksumEnabled", BooleanValue(true));

    //LogComponentEnable("TapBridge",LOG_LEVEL_ALL);
    //LogComponentEnable("CsmaHelper",LOG_LEVEL_ALL);
    //LogComponentEnable("NrUeNetDevice",LOG_LEVEL_ALL);
    //LogComponentEnable("NrGnbNetDevice",LOG_LEVEL_ALL);
    //LogComponentEnable("Node",LOG_LEVEL_FUNCTION);
    LogComponentEnable("EpcPgwApplication", LOG_LEVEL_INFO);




    uint16_t gnbUeDistance = 5; // meters

    Time simTime = Seconds(360);

    uint16_t numerology = 0;
    double centralFrequency = 3.5e9;
    double bandwidth = 400e6;
    double gnbTxPower = 30; // dBm
    double ueTxPower = 53;  // dBm

    int64_t randomStream = 1;


    NodeContainer gnbContainer;
    gnbContainer.Create(1);
    Names::Add("gNB",gnbContainer.Get(0));
    NodeContainer ueContainer;
    ueContainer.Create(1);
    Names::Add("Ue0",ueContainer.Get(0));

    MobilityHelper mobility;
    mobility.SetMobilityModel("ns3::ConstantPositionMobilityModel");
    Ptr<ListPositionAllocator> positionAllocUe = CreateObject<ListPositionAllocator>();
    positionAllocUe->Add(Vector(0.0, 0.0, 10.0));
    positionAllocUe->Add(Vector(gnbUeDistance, 0.0, 1.5));
    mobility.SetPositionAllocator(positionAllocUe);
    mobility.Install(gnbContainer);
    mobility.Install(ueContainer.Get(0));

    /* The default topology is the following:
     *
     *                    UE (20.0, 0.0, 1.5)
     *                   .
     *                  .
     *                 .
     *             (20 m)     
     *              .
     *             . 
     *            .
     *         gNB
     *   (0.0, 0.0, 10.0)               
     */

    /*
     * Setup the NR module. We create the various helpers needed for the
     * NR simulation:
     * - EpcHelper, which will setup the core network
     * - IdealBeamformingHelper, which takes care of the beamforming part
     * - NrHelper, which takes care of creating and connecting the various
     * part of the NR stack
     */
    Ptr<NrPointToPointEpcHelper> epcHelper = CreateObject<NrPointToPointEpcHelper>();
    Ptr<IdealBeamformingHelper> idealBeamformingHelper = CreateObject<IdealBeamformingHelper>();
    Ptr<NrHelper> nrHelper = CreateObject<NrHelper>();

    nrHelper->SetBeamformingHelper(idealBeamformingHelper);
    nrHelper->SetEpcHelper(epcHelper);

    CcBwpCreator ccBwpCreator;
    const uint8_t numCcPerBand = 1; 

    BandwidthPartInfoPtrVector allBwps;
    CcBwpCreator::SimpleOperationBandConf bandConf(centralFrequency,
                                                   bandwidth,
                                                   numCcPerBand,
                                                   BandwidthPartInfo::UMi_StreetCanyon);

    OperationBandInfo band = ccBwpCreator.CreateOperationBandContiguousCc(bandConf);

    /*
     * The configured spectrum division is:
     * ------------Band--------------
     * ------------CC1----------------
     * ------------BWP1---------------
     */

    Config::SetDefault("ns3::ThreeGppChannelModel::UpdatePeriod", TimeValue(MilliSeconds(0)));
    nrHelper->SetChannelConditionModelAttribute("UpdatePeriod", TimeValue(MilliSeconds(0)));
    nrHelper->SetPathlossAttribute("ShadowingEnabled", BooleanValue(false));

    nrHelper->InitializeOperationBand(&band);

    allBwps = CcBwpCreator::GetAllBwps({band});


    //Packet::EnableChecking();
    Packet::EnablePrinting();



    idealBeamformingHelper->SetAttribute("BeamformingMethod",
                                         TypeIdValue(DirectPathBeamforming::GetTypeId()));


    nrHelper->SetUeAntennaAttribute("NumRows", UintegerValue(2));
    nrHelper->SetUeAntennaAttribute("NumColumns", UintegerValue(4));
    nrHelper->SetUeAntennaAttribute("AntennaElement",
                                    PointerValue(CreateObject<IsotropicAntennaModel>()));



    nrHelper->SetGnbAntennaAttribute("NumRows", UintegerValue(4));
    nrHelper->SetGnbAntennaAttribute("NumColumns", UintegerValue(8));
    nrHelper->SetGnbAntennaAttribute("AntennaElement",
                                     PointerValue(CreateObject<ThreeGppAntennaModel>()));

    uint32_t bwpId = 0;

    nrHelper->SetGnbBwpManagerAlgorithmAttribute("NGBR_LOW_LAT_EMBB", UintegerValue(bwpId));

    nrHelper->SetUeBwpManagerAlgorithmAttribute("NGBR_LOW_LAT_EMBB", UintegerValue(bwpId));


    NetDeviceContainer enbNetDev = nrHelper->InstallGnbDevice(gnbContainer, allBwps);
    NetDeviceContainer ueNetDev = nrHelper->InstallUeDevice(ueContainer, allBwps);

    randomStream += nrHelper->AssignStreams(enbNetDev, randomStream);
    randomStream += nrHelper->AssignStreams(ueNetDev, randomStream);

    nrHelper->GetGnbPhy(enbNetDev.Get(0), 0)->SetAttribute("Numerology", UintegerValue(numerology));
    nrHelper->GetGnbPhy(enbNetDev.Get(0), 0)->SetAttribute("TxPower", DoubleValue(gnbTxPower));
    nrHelper->GetUePhy(ueNetDev.Get(0), 0)->SetAttribute("TxPower", DoubleValue(ueTxPower));


    for (auto it = enbNetDev.Begin(); it != enbNetDev.End(); ++it)
    {
        DynamicCast<NrGnbNetDevice>(*it)->UpdateConfig();
    }

    for (auto it = ueNetDev.Begin(); it != ueNetDev.End(); ++it)
    {
        DynamicCast<NrUeNetDevice>(*it)->UpdateConfig();
    }
    Ptr <Node> pgw = epcHelper->GetPgwNode();
    Names::Add("Packet Gateway",pgw);
    Ptr <Node> sgw = epcHelper->GetSgwNode();
    Names::Add("Services Gateway",sgw);
    //Ptr <Node> x = 
    InternetStackHelper internetStackHelper;

    Ipv4StaticRoutingHelper ipv4RoutingHelper;
    internetStackHelper.Install(ueContainer);

    Ipv4InterfaceContainer ueIpIface = epcHelper->AssignUeIpv4Address(NetDeviceContainer(ueNetDev));
    nrHelper->AttachToClosestEnb(ueNetDev, enbNetDev);

    CsmaHelper csma;
    csma.SetChannelAttribute("DataRate", DataRateValue(100000000000));
    csma.SetChannelAttribute("Delay", TimeValue(MilliSeconds(0)));
    Ipv4AddressHelper ipv4h;
    
    NodeContainer ghostNodes;
    ghostNodes.Create(2);
    Names::Add("ContainerUe",ghostNodes.Get(0));
    Names::Add("ContainergNB",ghostNodes.Get(1));
    //Names::Add("ContainerPgw",ghostNodes.Get(2));
    //Names::Add("ContainerSgw",ghostNodes.Get(3));
    internetStackHelper.Install(ghostNodes);

    NodeContainer LeftNodes;
    LeftNodes.Add(ueContainer.Get(0));
    LeftNodes.Add(ghostNodes.Get(0));


    NodeContainer RightNodes;
    RightNodes.Add(gnbContainer.Get(0));
    RightNodes.Add(ghostNodes.Get(1));


    NetDeviceContainer deviceLeft = csma.Install(LeftNodes);
    ipv4h.SetBase("10.1.1.0", "255.255.255.0","0.0.0.1");
    Ipv4InterfaceContainer interfacesLeft = ipv4h.Assign(deviceLeft);

    NetDeviceContainer deviceRight = csma.Install(RightNodes);
    ipv4h.SetBase("10.1.2.0","255.255.255.0","0.0.0.1");
    Ipv4InterfaceContainer interfacesRight = ipv4h.Assign(deviceRight);

    NodeList allnodes = NodeList();
    //Ipv4GlobalRoutingHelper::PopulateRoutingTables();

    for (uint32_t j = 0; j < ueContainer.GetN(); ++j)
    {
        Ptr<Ipv4StaticRouting> ueStaticRouting =
            ipv4RoutingHelper.GetStaticRouting(ueContainer.Get(j)->GetObject<Ipv4>());
        ueStaticRouting->SetDefaultRoute(epcHelper->GetUeDefaultGatewayAddress(), 1);
    }
    
    Ptr <Ipv4StaticRouting> clientroute = ipv4RoutingHelper.GetStaticRouting(ghostNodes.Get(0)->GetObject<Ipv4>());
    clientroute->AddNetworkRouteTo(Ipv4Address("10.0.0.0"),Ipv4Mask("255.0.0.0"),Ipv4Address("10.0.0.6"),1);
    clientroute->AddNetworkRouteTo(Ipv4Address("10.1.2.0"),Ipv4Mask("255.255.255.0"),1);
    clientroute->AddNetworkRouteTo(Ipv4Address("10.100.200.0"),Ipv4Mask("/24"),1);

    Ptr <Ipv4StaticRouting> ueroute = ipv4RoutingHelper.GetStaticRouting(ueContainer.Get(0)->GetObject<Ipv4>());
    ueroute->AddNetworkRouteTo(Ipv4Address("10.0.0.0"),Ipv4Mask("255.0.0.0"),Ipv4Address("7.0.0.1"),1);
    ueroute->AddNetworkRouteTo(Ipv4Address("10.1.1.0"),Ipv4Mask("255.255.255.0"),2);
    ueroute->AddNetworkRouteTo(Ipv4Address("10.1.2.0"),Ipv4Mask("255.255.255.0"),1);
    ueroute->AddNetworkRouteTo(Ipv4Address("10.100.200.0"),Ipv4Mask("/24"),1);


    Ptr <Ipv4StaticRouting> gnbroute = ipv4RoutingHelper.GetStaticRouting(gnbContainer.Get(0)->GetObject<Ipv4>());
    gnbroute->AddNetworkRouteTo(Ipv4Address("7.0.0.0"),Ipv4Mask("255.0.0.0"),Ipv4Address("7.0.0.1"),1);
    gnbroute->AddNetworkRouteTo(Ipv4Address("10.1.1.0"),Ipv4Mask("255.255.255.0"),1);
    gnbroute->AddNetworkRouteTo(Ipv4Address("10.1.2.0"),Ipv4Mask("255.255.255.0"),2);
    gnbroute->AddNetworkRouteTo(Ipv4Address("10.100.200.0"),Ipv4Mask("/24"),2);

    Ptr <Ipv4StaticRouting> bs_container_route = ipv4RoutingHelper.GetStaticRouting(ghostNodes.Get(1)->GetObject<Ipv4>());
    bs_container_route->AddNetworkRouteTo(Ipv4Address("10.1.2.0"),Ipv4Mask("255.255.255.0"),2);
    bs_container_route->AddNetworkRouteTo(Ipv4Address("7.0.0.0"),Ipv4Mask("255.0.0.0"),1);

    Ptr <Ipv4StaticRouting> pgwroute = ipv4RoutingHelper.GetStaticRouting(pgw->GetObject<Ipv4>());
    pgwroute->AddNetworkRouteTo(Ipv4Address("10.0.0.0"),Ipv4Mask("255.0.0.0"),2);
    pgwroute->AddNetworkRouteTo(Ipv4Address("10.1.1.0"),Ipv4Mask("255.255.255.0"),1);
    pgwroute->AddNetworkRouteTo(Ipv4Address("10.1.2.0"),Ipv4Mask("255.255.255.0"),2);
    pgwroute->AddNetworkRouteTo(Ipv4Address("10.100.200.0"),Ipv4Mask("/24"),2);

    Ptr<Ipv4StaticRouting> sgwroute = ipv4RoutingHelper.GetStaticRouting(sgw->GetObject<Ipv4>());
    sgwroute->AddNetworkRouteTo(Ipv4Address("10.1.1.0"),Ipv4Mask("255.255.255.0"),1);
    sgwroute->AddNetworkRouteTo(Ipv4Address("10.1.2.0"),Ipv4Mask("255.255.255.0"),3);
    sgwroute->AddNetworkRouteTo(Ipv4Address("10.100.200.0"),Ipv4Mask("255.255.255.0"),3);
    



    addBackAddress(pgw,ueNetDev.Get(0),Ipv4Address("10.1.1.2"));

    TapBridgeHelper tapBridge;
    tapBridge.SetAttribute("Mode",StringValue("UseBridge"));
    tapBridge.SetAttribute("DeviceName", StringValue("tap-ue"));
    tapBridge.Install(LeftNodes.Get(1), deviceLeft.Get(1));
    tapBridge.SetAttribute("DeviceName", StringValue("tap-gnb"));
    tapBridge.Install(RightNodes.Get(1), deviceRight.Get(1));


    csma.EnablePcapAll("output",false);

    Simulator::Schedule(MilliSeconds(0.5), &ipv4, allnodes);
    Simulator::Schedule(MilliSeconds(1),&MacAddress, enbNetDev, ueNetDev);
    Simulator::Schedule(MilliSeconds(1.2),&PrintRoutingTable,allnodes);
    Simulator::Stop(Seconds(600000));
    Simulator::Run();
    Simulator::Destroy();


    return 0;
}

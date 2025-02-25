include <BOSL2/std.scad>
$fn=24;

wallWidth = 4;

width = 180;
depth = 180;
height = 20;


funnelYOffset = 60; 

funnelLocation = [0,12,-300];
funnelWidth = 220;
funnelLength = 95;
funnelHeight= 320;
funnelRotate = 10;

funnelBlockerSize =  [width-wallWidth+40,depth*3,height+wallWidth+280];
funnelBlockerLocaton = [0,-25,wallWidth-80];


boxLocation = [0,0,funnelHeight-100];
coverHeight = 14;
coverLocation = [wallWidth+3,0,155];

holderHeightOffset = 140;
holderXOffset = 30;
holderLocation = coverLocation+[-wallWidth,holderXOffset,holderHeightOffset];
holderSize = [depth+wallWidth,85,90];
debug = true;



// part = "funnel|cover|box-sizer"
part = "box";


allBottomEdges = [BOTTOM+FRONT, BOTTOM+BACK, BOTTOM+LEFT, BOTTOM+RIGHT, 
           LEFT+FRONT, LEFT+BACK, RIGHT+FRONT, RIGHT+BACK];

if(part == "boxv1"){
     difference(){          
        cuboid(
            [width,depth,height], rounding=5,
            edges=allBottomEdges,
        );

        translate([0,0,wallWidth])
        #cuboid(
            [width-wallWidth,depth-wallWidth,height], rounding=5,
            edges=allBottomEdges,
            
        );
    }
}



if(part == "funnel"){

rotate([180,0,0])
//translate(funnelLocation) //[0,funnelYOffset,-funnelHeight+3])
difference(){
    rotate([funnelRotate,0,0])
    translate(funnelLocation)
    prismoid([funnelWidth,funnelLength], [0,0], h=funnelHeight);
        
    

    rotate([funnelRotate,0,0])
     translate(funnelLocation+[0,0,-wallWidth])
    prismoid([funnelWidth-wallWidth,funnelLength-wallWidth], [0,0], h=funnelHeight);
    
    //translate(boxLocation)
    translate(funnelBlockerLocaton)
    cuboid(
       funnelBlockerSize, rounding=5,
        edges=allBottomEdges,
    );
    
      if (debug){
            /// debug sizing cube
            translate([100,-350,-370])
            cuboid([1000,1000,280], anchor=CENTER);
            }
        }
    }
   
   if(part == "box"){
    
  difference(){
    translate(boxLocation)
         cuboid(
        [width+wallWidth,depth+wallWidth,height+wallWidth+7], rounding=1,
        edges=allBottomEdges,
        
    );
    translate(boxLocation)
    translate([0,0,wallWidth])
    cuboid(
        [width-wallWidth,depth-wallWidth,height+wallWidth+300], rounding=5,
        edges=allBottomEdges,
    );
    // cover-cut out
   translate(boxLocation+[0,5,1])
    cuboid([width-wallWidth-5,depth-wallWidth+10,coverHeight], anchor=CENTER);
    
       translate(boxLocation+[0,5,1])
    #cuboid([width-wallWidth+2,depth-wallWidth+3,coverHeight], anchor=CENTER);
    
 
     translate(boxLocation+[0,10,-6])
    cuboid([width-wallWidth-10,depth-wallWidth-10,coverHeight+30], anchor=CENTER);
    
}
}

if(part == "holder-box" || part == "all"){
    translate(holderLocation)
    difference(){
        cuboid(holderSize, rounding=3, anchor=CENTER);
        translate([0,0,10])
        cuboid(holderSize+[-wallWidth,-wallWidth,10], rounding=3, anchor=CENTER);
       }
       }

if(part == "cover"){
    cuboid([width-wallWidth+10,depth-wallWidth,coverHeight], anchor=CENTER);
    
  }


include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 20;

clipStyle = "clip";

/*
holderLength = 50;
holeDepth = 25;
holderRadius = 6;

holeRadius = 4;
*/
clipInwardPush = 7; // 
clipStyle = "noClip";

designFileName = "small_handle_extender";
holderLength = 15;

holderRadius = 14;
holeDepth = 14;
holeRadius = 6.5;
clipDownOffset = 6; //
iName = "small-stubby-holder";
version = "v0.5";
wedgeOut = 4; //

clipDownOffset = 6;
clipInwardPush = 7;
clipStyle = "clip";
designFileName = "small_handle_extender";
holderLength = 25;
holderRadius = 14;
holeDepth = 14;
holeRadius = 6.5;
iName = "small-stubby-holder";
version = "v0.5";
wedgeOut = 4;



module hex(radius=5, height=10, holeRadius=8, holeDepth=1, textureStr="rough"){
difference(){
    // Define a hexagon base
    hex_path = hexagon(r=radius, rounding=1); 

    // Load your desired texture
    tex = texture("rough");
    if (len(textureStr) > 0){
    // Extrude the hexagon with the texture on the top face
        linear_sweep(hex_path, height=height, texture=tex, tex_size=[15, 15]);
    } else {
        linear_sweep(hex_path, height=height, tex_size=[15, 15]);
    }
    
    
    up(height-holeDepth)
    linear_sweep(hexagon(ir=holeRadius), height=holeDepth+0.1);
    
    
        yrot(180)
        chamfer_cylinder_mask(r=radius, chamfer=1);
    }
    
   }
   
   
clipDepth = 2;
clipWidth = 5;
clipHeight = 8;

clipSize = [clipWidth, clipDepth, clipHeight];
clipOffset = 1;
   clipConnectorWidth = 10;
   clipCenterOffset = holderRadius-clipInwardPush;
   
   clipConnectorHeight = 15;
clipUp = holderLength-clipDownOffset;
clipConnectorUp = clipHeight-clipConnectorHeight;

module clip(){
    /*up(clipConnectorUp)
    fwd(clipConnectorWidth/2)
    cuboid([widerPipeRadius,clipConnectorWidth,clipConnectorHeight], rounding=1,  edges="Y", anchor=BOTTOM);
*/
    difference(){
    cuboid(clipSize, chamfer=0.5, edges="Y", anchor=BOTTOM)
    position(TOP){
    up(0.5)
    down(wedgeOut/2)
  
    xrot(270)
    back(0.5)
    scale([1,1,1.5])
        top_half()
        wedge(wedgeOut,30);
        }
   }
   }
   
   collarholeMultipler = 0.9;

module small_handle_extender(clipStyle=clipStyle){
    hex(radius=holderRadius, height=holderLength, holeRadius=holeRadius, holeDepth=holeDepth);
//    attach(TOP){
up(holderLength)
        hex(radius=holeRadius*1.2, holeRadius=holeRadius*collarholeMultipler, height=1, holeDepth=30, textureStr="");
  //  }
    if(clipStyle == "clip"){
    fwd(clipCenterOffset)
        up(clipUp)
        clip();
        }
 }


small_handle_extender(clipStyle=clipStyle);
